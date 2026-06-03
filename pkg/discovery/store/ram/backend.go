package ram

import (
	"context"
	"sync"

	"github.com/sora-soft/sora-go-framework/pkg/discovery"
)

type RamBackend struct {
	store     *store
	registry  *ramRegistry
	discovery *ramDiscovery

	electionsMu sync.Mutex
	elections   map[string]*ramElection
}

func NewRamBackend() *RamBackend {
	s := newStore()
	return &RamBackend{
		store:     s,
		registry:  &ramRegistry{store: s},
		discovery: &ramDiscovery{store: s},
		elections: make(map[string]*ramElection),
	}
}

func (b *RamBackend) Connect(_ context.Context) error {
	return nil
}

func (b *RamBackend) Disconnect() error {
	return nil
}

func (b *RamBackend) Registry() discovery.Registry {
	return b.registry
}

func (b *RamBackend) Discovery() discovery.Discovery {
	return b.discovery
}

func (b *RamBackend) NewElection(name string) discovery.Election {
	b.electionsMu.Lock()
	defer b.electionsMu.Unlock()

	if e, ok := b.elections[name]; ok {
		return e
	}

	e := &ramElection{name: name}
	b.elections[name] = e
	return e
}

func (b *RamBackend) GetInfo() discovery.BackendInfo {
	return discovery.BackendInfo{
		Type:    "ram",
		Version: "0.0.0",
	}
}
