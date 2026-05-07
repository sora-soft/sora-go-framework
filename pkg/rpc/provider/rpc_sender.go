package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
)

type RpcSender struct {
	endpoint       discovery.EndpointMeta
	conn           *rpc.Connection
	connMu         sync.RWMutex
	pending        map[string]chan packet.Packet
	pendingMu      sync.RWMutex
	codec          rpc.Codec
	transportConf  rpc.TransportConfig
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewRpcSender(endpoint discovery.EndpointMeta, codec rpc.Codec, transportConf rpc.TransportConfig) *RpcSender {
	return &RpcSender{
		endpoint:      endpoint,
		pending:       make(map[string]chan packet.Packet),
		codec:         codec,
		transportConf: transportConf,
	}
}

func (s *RpcSender) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	go s.connectLoop()
}

func (s *RpcSender) connectLoop() {
	delay := 500 * time.Millisecond
	for {
		transport := s.transportConf.Factory()
		conn := rpc.NewConnection(transport, s.transportConf.Options)
		conn.OnResponse = s.handleResponse

		err := conn.Start(s.ctx, rpc.ListenerInfo{
			Protocol: s.endpoint.Protocol,
			Endpoint: s.endpoint.Endpoint,
			Codecs:   s.endpoint.Codecs,
		}, s.codec)

		if err == nil {
			delay = 500 * time.Millisecond
			s.connMu.Lock()
			s.conn = conn
			s.connMu.Unlock()

			stateCh := conn.LifeCycle.Listen()
			for state := range stateCh {
				if state == rpc.ConnectorStateError || state == rpc.ConnectorStateStopped {
					s.connMu.Lock()
					s.conn = nil
					s.connMu.Unlock()
					conn.LifeCycle.RemoveListen(stateCh)
					break
				}
			}
		}

		s.failAllPending()

		delay = min(delay*2, 10*time.Second)
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func (s *RpcSender) handleResponse(_ *rpc.Connection, pkt packet.Packet) {
	s.pendingMu.RLock()
	requestId, ok := pkt.Headers[rpc.HeaderRpcId]
	if !ok {
		s.pendingMu.RUnlock()
		return
	}
	ch, exists := s.pending[requestId]
	s.pendingMu.RUnlock()

	if exists {
		s.pendingMu.Lock()
		delete(s.pending, requestId)
		s.pendingMu.Unlock()
		ch <- pkt
	}
}

func (s *RpcSender) callRpcRaw(ctx context.Context, method string, payload []byte, headers map[string]string) (packet.Packet, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return packet.Packet{}, err
	}
	requestId := hex.EncodeToString(b)

	if headers == nil {
		headers = make(map[string]string)
	}
	headers[rpc.HeaderRpcId] = requestId

	ch := make(chan packet.Packet, 1)
	s.pendingMu.Lock()
	s.pending[requestId] = ch
	s.pendingMu.Unlock()

	pkt := packet.NewDecodedPacket(packet.PacketOpcodeRequest, method, "", headers, payload, s.codec)

	s.connMu.RLock()
	conn := s.conn
	s.connMu.RUnlock()

	if conn == nil {
		s.pendingMu.Lock()
		delete(s.pending, requestId)
		s.pendingMu.Unlock()
		return packet.Packet{}, ErrConnectionLost
	}

	if err := conn.SendRaw(ctx, pkt); err != nil {
		s.pendingMu.Lock()
		delete(s.pending, requestId)
		s.pendingMu.Unlock()
		return packet.Packet{}, err
	}

	select {
	case res := <-ch:
		return res, nil
	case <-ctx.Done():
		s.pendingMu.Lock()
		delete(s.pending, requestId)
		s.pendingMu.Unlock()
		return packet.Packet{}, ctx.Err()
	case <-s.ctx.Done():
		s.pendingMu.Lock()
		delete(s.pending, requestId)
		s.pendingMu.Unlock()
		return packet.Packet{}, ErrSenderStopped
	}
}

func (s *RpcSender) sendNotifyRaw(ctx context.Context, method string, payload []byte, headers map[string]string) error {
	pkt := packet.NewDecodedPacket(packet.PacketOpcodeNotify, method, "", headers, payload, s.codec)

	s.connMu.RLock()
	conn := s.conn
	s.connMu.RUnlock()

	if conn == nil {
		return ErrConnectionLost
	}

	return conn.SendRaw(ctx, pkt)
}

func (s *RpcSender) failAllPending() {
	s.pendingMu.Lock()
	pending := s.pending
	s.pending = make(map[string]chan packet.Packet)
	s.pendingMu.Unlock()

	for _, ch := range pending {
		close(ch)
	}
}

func (s *RpcSender) Destroy() {
	if s.cancel != nil {
		s.cancel()
	}
	s.failAllPending()

	s.connMu.Lock()
	conn := s.conn
	s.conn = nil
	s.connMu.Unlock()

	if conn != nil {
		conn.Disconnect()
	}
}

func (s *RpcSender) isReady() bool {
	s.connMu.RLock()
	conn := s.conn
	s.connMu.RUnlock()
	return conn != nil && conn.LifeCycle.GetState() == rpc.ConnectorStateReady
}
