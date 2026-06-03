package rpc

import (
	"sync"
	"time"

	"github.com/sora-soft/sora-go-framework/pkg/utility/errorx"
)

type Pinger struct {
	mu       sync.Mutex
	pending  map[int]chan error
	nextID   int
	sendPing func(id int) error
	onError  func(err error)
	state    func() ConnectorState
	opts     ConnectorPingOptions
	stop     chan struct{}
}

func NewPinger(sendPing func(int) error, onError func(error), state func() ConnectorState, opts ConnectorPingOptions) *Pinger {
	if opts.Interval == 0 {
		opts.Interval = 10 * time.Second
	}
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}
	return &Pinger{
		pending:  make(map[int]chan error),
		sendPing: sendPing,
		onError:  onError,
		state:    state,
		opts:     opts,
	}
}

func (p *Pinger) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stop != nil || !p.opts.Enabled {
		return
	}
	stop := make(chan struct{})
	p.stop = stop

	go func() {
		ticker := time.NewTicker(p.opts.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if !p.opts.Enabled || p.state() != ConnectorStateReady {
					continue
				}
				p.pingOnce()
			}
		}
	}()
}

func (p *Pinger) Stop() {
	p.mu.Lock()
	if p.stop != nil {
		close(p.stop)
		p.stop = nil
	}
	for _, ch := range p.pending {
		close(ch)
	}
	p.pending = make(map[int]chan error)
	p.mu.Unlock()
}

func (p *Pinger) pingOnce() {
	ch := make(chan error, 1)
	id := p.addPending(ch)
	defer p.removePending(id)

	if err := p.sendPing(id); err != nil {
		return
	}

	select {
	case err, ok := <-ch:
		if ok && err != nil {
			p.onError(err)
		}
	case <-time.After(p.opts.Timeout):
		p.onError(errorx.New("ERR_PING_PONG_TIMEOUT", errorx.LevelUnexpected, "PingerError", "pong timeout", map[string]any{"pingId": id}))
	}
}

func (p *Pinger) OnPong(id int, err error) {
	p.mu.Lock()
	ch, ok := p.pending[id]
	if ok {
		delete(p.pending, id)
	}
	p.mu.Unlock()
	if ok {
		ch <- err
	}
}

func (p *Pinger) addPending(ch chan error) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nextID++
	id := p.nextID
	p.pending[id] = ch
	return id
}

func (p *Pinger) removePending(id int) {
	p.mu.Lock()
	delete(p.pending, id)
	p.mu.Unlock()
}
