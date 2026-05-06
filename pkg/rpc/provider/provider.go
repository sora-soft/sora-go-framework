package provider

import (
	"context"
	"math/rand"
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
)

type ProviderOptions struct {
	LabelFilter *LabelFilter
}

type Provider struct {
	serviceName  string
	disco        discovery.Discovery
	labelFilter  *LabelFilter
	senders      map[string]*RpcSender
	sendersBySvc map[string][]*RpcSender
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewProvider(serviceName string, disco discovery.Discovery, opts ProviderOptions) *Provider {
	return &Provider{
		serviceName:  serviceName,
		disco:        disco,
		labelFilter:  opts.LabelFilter,
		senders:      make(map[string]*RpcSender),
		sendersBySvc: make(map[string][]*RpcSender),
	}
}

func (p *Provider) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	go p.watchLoop()
}

func (p *Provider) watchLoop() {
	endpointCh := p.disco.WatchEndpoints(p.ctx)
	for snapshot := range endpointCh {
		filtered := make([]discovery.EndpointMeta, 0, len(snapshot))
		for _, ep := range snapshot {
			if ep.TargetName == p.serviceName && p.labelFilter.IsSatisfy(ep.Labels) {
				filtered = append(filtered, ep)
			}
		}

		p.mu.Lock()
		currentIds := make(map[string]bool)
		for id := range p.senders {
			currentIds[id] = true
		}

		filteredIds := make(map[string]bool)
		for _, ep := range filtered {
			filteredIds[ep.ID] = true
		}

		for _, ep := range filtered {
			if !currentIds[ep.ID] {
				p.addSenderLocked(p.ctx, ep)
			}
		}

		for id := range currentIds {
			if !filteredIds[id] {
				p.removeSenderLocked(id)
			}
		}
		p.mu.Unlock()
	}
}

func (p *Provider) addSenderLocked(ctx context.Context, endpoint discovery.EndpointMeta) {
	var codec rpc.Codec
	for _, code := range endpoint.Codecs {
		c, ok := rpc.GetCodec(code)
		if ok {
			codec = c
			break
		}
	}
	if codec == nil {
		return
	}

	conf, ok := rpc.GetTransportConfig(endpoint.Protocol)
	if !ok {
		return
	}

	sender := NewRpcSender(endpoint, p, codec, conf)
	sender.Start(ctx)

	p.senders[endpoint.ID] = sender
	p.sendersBySvc[endpoint.TargetID] = append(p.sendersBySvc[endpoint.TargetID], sender)
}

func (p *Provider) removeSenderLocked(endpointId string) {
	sender, ok := p.senders[endpointId]
	if !ok {
		return
	}

	delete(p.senders, endpointId)

	svcId := sender.endpoint.TargetID
	senders := p.sendersBySvc[svcId]
	for i, s := range senders {
		if s == sender {
			p.sendersBySvc[svcId] = append(senders[:i], senders[i+1:]...)
			break
		}
	}
	if len(p.sendersBySvc[svcId]) == 0 {
		delete(p.sendersBySvc, svcId)
	}

	sender.Destroy()
}

func (p *Provider) selectSender() (*RpcSender, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var ready []*RpcSender
	var totalWeight int
	for _, s := range p.senders {
		if s.isReady() {
			ready = append(ready, s)
			totalWeight += s.endpoint.Weight
		}
	}
	if len(ready) == 0 {
		return nil, ErrNoAvailableEndpoint
	}

	target := rand.Intn(totalWeight)
	accum := 0
	for _, s := range ready {
		accum += s.endpoint.Weight
		if target < accum {
			return s, nil
		}
	}
	return ready[0], nil
}

func (p *Provider) selectSenderByService(serviceId string) (*RpcSender, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	senders, ok := p.sendersBySvc[serviceId]
	if !ok {
		return nil, ErrServiceNotFound
	}
	for _, s := range senders {
		if s.isReady() {
			return s, nil
		}
	}
	return nil, ErrServiceNotFound
}

func (p *Provider) CallRpc(ctx context.Context, method string, req any, opts ...CallOption) (packet.Packet, error) {
	if p.ctx != nil {
		select {
		case <-p.ctx.Done():
			return packet.Packet{}, ErrProviderStopped
		default:
		}
	}

	options := defaultCallOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var sender *RpcSender
	var err error
	if options.TargetID != "" {
		sender, err = p.selectSenderByService(options.TargetID)
	} else {
		sender, err = p.selectSender()
	}
	if err != nil {
		return packet.Packet{}, err
	}

	payload, err := sender.codec.Marshal(req)
	if err != nil {
		return packet.Packet{}, err
	}

	callCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	return sender.callRpcRaw(callCtx, method, payload, nil)
}

func (p *Provider) Stop() {
	if p.cancel != nil {
		p.cancel()
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for id, sender := range p.senders {
		sender.Destroy()
		delete(p.senders, id)
	}
	p.sendersBySvc = make(map[string][]*RpcSender)
}
