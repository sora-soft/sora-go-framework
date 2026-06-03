package ram

import (
	"context"
	"sync"

	"github.com/sora-soft/sora-go-framework/pkg/discovery"
)

type watcher[T any] struct {
	ch chan []T
}

type notifier[T any] struct {
	mu       sync.RWMutex
	watchers []*watcher[T]
}

func (n *notifier[T]) subscribe() *watcher[T] {
	w := &watcher[T]{ch: make(chan []T, 8)}
	n.mu.Lock()
	n.watchers = append(n.watchers, w)
	n.mu.Unlock()
	return w
}

func (n *notifier[T]) unsubscribe(w *watcher[T]) {
	n.mu.Lock()
	for i, ww := range n.watchers {
		if ww == w {
			n.watchers = append(n.watchers[:i], n.watchers[i+1:]...)
			break
		}
	}
	n.mu.Unlock()
}

func (n *notifier[T]) push(snapshot []T) {
	n.mu.RLock()
	for _, w := range n.watchers {
		select {
		case w.ch <- snapshot:
		default:
		}
	}
	n.mu.RUnlock()
}

type store struct {
	mu sync.RWMutex

	nodes     map[string]discovery.NodeMeta
	services  map[string]discovery.ServiceMeta
	workers   map[string]discovery.WorkerMeta
	endpoints map[string]discovery.EndpointMeta

	nodeNotifier     notifier[discovery.NodeMeta]
	serviceNotifier  notifier[discovery.ServiceMeta]
	workerNotifier   notifier[discovery.WorkerMeta]
	endpointNotifier notifier[discovery.EndpointMeta]
}

func newStore() *store {
	return &store{
		nodes:     make(map[string]discovery.NodeMeta),
		services:  make(map[string]discovery.ServiceMeta),
		workers:   make(map[string]discovery.WorkerMeta),
		endpoints: make(map[string]discovery.EndpointMeta),
	}
}

func (s *store) snapshotNodes() []discovery.NodeMeta {
	s.mu.RLock()
	result := make([]discovery.NodeMeta, 0, len(s.nodes))
	for _, n := range s.nodes {
		result = append(result, n)
	}
	s.mu.RUnlock()
	return result
}

func (s *store) snapshotServices() []discovery.ServiceMeta {
	s.mu.RLock()
	result := make([]discovery.ServiceMeta, 0, len(s.services))
	for _, svc := range s.services {
		result = append(result, svc)
	}
	s.mu.RUnlock()
	return result
}

func (s *store) snapshotWorkers() []discovery.WorkerMeta {
	s.mu.RLock()
	result := make([]discovery.WorkerMeta, 0, len(s.workers))
	for _, w := range s.workers {
		result = append(result, w)
	}
	s.mu.RUnlock()
	return result
}

func (s *store) snapshotEndpoints() []discovery.EndpointMeta {
	s.mu.RLock()
	result := make([]discovery.EndpointMeta, 0, len(s.endpoints))
	for _, ep := range s.endpoints {
		result = append(result, ep)
	}
	s.mu.RUnlock()
	return result
}

func watchEntities[T any](ctx context.Context, n *notifier[T], snapshotFn func() []T) <-chan []T {
	w := n.subscribe()
	ch := make(chan []T, 8)

	initial := snapshotFn()
	ch <- initial

	go func() {
		defer close(ch)
		defer n.unsubscribe(w)

		for {
			select {
			case <-ctx.Done():
				return
			case snapshot, ok := <-w.ch:
				if !ok {
					return
				}
				select {
				case ch <- snapshot:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch
}
