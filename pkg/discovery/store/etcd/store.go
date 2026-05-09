package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	"go.etcd.io/etcd/api/v3/mvccpb"
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

type entityType int

const (
	entityNode     entityType = 0
	entityService  entityType = 1
	entityWorker   entityType = 2
	entityEndpoint entityType = 3
)

type revisionMeta struct {
	createRevision int64
	modRevision    int64
}

type store struct {
	mu sync.RWMutex

	nodes     map[string]discovery.NodeMeta
	services  map[string]discovery.ServiceMeta
	workers   map[string]discovery.WorkerMeta
	endpoints map[string]discovery.EndpointMeta

	nodeRevisions     map[string]revisionMeta
	serviceRevisions  map[string]revisionMeta
	workerRevisions   map[string]revisionMeta
	endpointRevisions map[string]revisionMeta

	nodeNotifier     notifier[discovery.NodeMeta]
	serviceNotifier  notifier[discovery.ServiceMeta]
	workerNotifier   notifier[discovery.WorkerMeta]
	endpointNotifier notifier[discovery.EndpointMeta]
}

func newStore() *store {
	return &store{
		nodes:             make(map[string]discovery.NodeMeta),
		services:          make(map[string]discovery.ServiceMeta),
		workers:           make(map[string]discovery.WorkerMeta),
		endpoints:         make(map[string]discovery.EndpointMeta),
		nodeRevisions:     make(map[string]revisionMeta),
		serviceRevisions:  make(map[string]revisionMeta),
		workerRevisions:   make(map[string]revisionMeta),
		endpointRevisions: make(map[string]revisionMeta),
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

func entityPrefix(prefix string, et entityType) string {
	switch et {
	case entityNode:
		return prefix + "/node"
	case entityService:
		return prefix + "/service"
	case entityWorker:
		return prefix + "/worker"
	case entityEndpoint:
		return prefix + "/endpoint"
	default:
		panic(fmt.Sprintf("unknown entity type: %d", et))
	}
}

func entityKey(prefix string, et entityType, id string) string {
	return entityPrefix(prefix, et) + "/" + id
}

func (s *store) updateNode(kv *mvccpb.KeyValue) {
	id := string(kv.Key)
	var meta discovery.NodeMeta
	if err := json.Unmarshal(kv.Value, &meta); err != nil {
		runtime.RT.FrameLogger.Error("EtcdStore", err, "failed to unmarshal node meta")
		return
	}

	s.mu.Lock()
	if existing, ok := s.nodeRevisions[id]; ok && existing.modRevision >= kv.ModRevision {
		s.mu.Unlock()
		return
	}
	s.nodes[id] = meta
	s.nodeRevisions[id] = revisionMeta{
		createRevision: kv.CreateRevision,
		modRevision:    kv.ModRevision,
	}
	s.mu.Unlock()

	s.nodeNotifier.push(s.snapshotNodes())
}

func (s *store) updateService(kv *mvccpb.KeyValue) {
	id := string(kv.Key)
	var meta discovery.ServiceMeta
	if err := json.Unmarshal(kv.Value, &meta); err != nil {
		runtime.RT.FrameLogger.Error("EtcdStore", err, "failed to unmarshal service meta")
		return
	}

	s.mu.Lock()
	if existing, ok := s.serviceRevisions[id]; ok && existing.modRevision >= kv.ModRevision {
		s.mu.Unlock()
		return
	}
	s.services[id] = meta
	s.serviceRevisions[id] = revisionMeta{
		createRevision: kv.CreateRevision,
		modRevision:    kv.ModRevision,
	}
	s.mu.Unlock()

	s.serviceNotifier.push(s.snapshotServices())
}

func (s *store) updateWorker(kv *mvccpb.KeyValue) {
	id := string(kv.Key)
	var meta discovery.WorkerMeta
	if err := json.Unmarshal(kv.Value, &meta); err != nil {
		return
	}

	s.mu.Lock()
	if existing, ok := s.workerRevisions[id]; ok && existing.modRevision >= kv.ModRevision {
		s.mu.Unlock()
		return
	}
	s.workers[id] = meta
	s.workerRevisions[id] = revisionMeta{
		createRevision: kv.CreateRevision,
		modRevision:    kv.ModRevision,
	}
	s.mu.Unlock()

	s.workerNotifier.push(s.snapshotWorkers())
}

func (s *store) updateEndpoint(kv *mvccpb.KeyValue) {
	id := string(kv.Key)
	var meta discovery.EndpointMeta
	if err := json.Unmarshal(kv.Value, &meta); err != nil {
		return
	}

	s.mu.Lock()
	if _, ok := s.services[meta.TargetID]; !ok {
		s.mu.Unlock()
		return
	}
	if existing, ok := s.endpointRevisions[id]; ok && existing.modRevision >= kv.ModRevision {
		s.mu.Unlock()
		return
	}
	s.endpoints[id] = meta
	s.endpointRevisions[id] = revisionMeta{
		createRevision: kv.CreateRevision,
		modRevision:    kv.ModRevision,
	}
	s.mu.Unlock()

	s.endpointNotifier.push(s.snapshotEndpoints())
}

func (s *store) deleteNode(id string) {
	s.mu.Lock()
	if _, ok := s.nodes[id]; !ok {
		s.mu.Unlock()
		return
	}
	delete(s.nodes, id)
	delete(s.nodeRevisions, id)
	s.mu.Unlock()

	s.nodeNotifier.push(s.snapshotNodes())
}

func (s *store) deleteService(id string) {
	s.mu.Lock()
	if _, ok := s.services[id]; !ok {
		s.mu.Unlock()
		return
	}
	delete(s.services, id)
	delete(s.serviceRevisions, id)
	s.mu.Unlock()

	s.serviceNotifier.push(s.snapshotServices())
}

func (s *store) deleteWorker(id string) {
	s.mu.Lock()
	if _, ok := s.workers[id]; !ok {
		s.mu.Unlock()
		return
	}
	delete(s.workers, id)
	delete(s.workerRevisions, id)
	s.mu.Unlock()

	s.workerNotifier.push(s.snapshotWorkers())
}

func (s *store) deleteEndpoint(id string) {
	s.mu.Lock()
	if _, ok := s.endpoints[id]; !ok {
		s.mu.Unlock()
		return
	}
	delete(s.endpoints, id)
	delete(s.endpointRevisions, id)
	s.mu.Unlock()

	s.endpointNotifier.push(s.snapshotEndpoints())
}
