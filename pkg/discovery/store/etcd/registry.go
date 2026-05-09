package etcd

import (
	"context"
	"sync"

	etcdcomp "github.com/sora-soft/sora-go-framework.git/pkg/component/etcd"
	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdRegistry struct {
	backend        *EtcdBackend
	etcdComp       *etcdcomp.EtcdComponent
	mu             sync.Mutex
	localNodes     map[string]discovery.NodeMeta
	localServices  map[string]discovery.ServiceMeta
	localWorkers   map[string]discovery.WorkerMeta
	localEndpoints map[string]discovery.EndpointMeta
}

func newEtcdRegistry(backend *EtcdBackend, impl *etcdcomp.EtcdComponent) *etcdRegistry {
	return &etcdRegistry{
		backend:        backend,
		etcdComp:       impl,
		localNodes:     make(map[string]discovery.NodeMeta),
		localServices:  make(map[string]discovery.ServiceMeta),
		localWorkers:   make(map[string]discovery.WorkerMeta),
		localEndpoints: make(map[string]discovery.EndpointMeta),
	}
}

func (r *etcdRegistry) RegisterNode(ctx context.Context, node discovery.NodeMeta) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityNode, node.ID)
	leaseID := r.etcdComp.LeaseID()
	if err := r.backend.putWithLease(ctx, key, node, leaseID); err != nil {
		return err
	}
	r.localNodes[node.ID] = node
	return nil
}

func (r *etcdRegistry) UnregisterNode(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityNode, id)
	if err := r.backend.deleteKey(ctx, key); err != nil {
		return err
	}
	delete(r.localNodes, id)
	return nil
}

func (r *etcdRegistry) RegisterService(ctx context.Context, service discovery.ServiceMeta) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityService, service.ID)
	leaseID := r.etcdComp.LeaseID()
	if err := r.backend.putWithLease(ctx, key, service, leaseID); err != nil {
		return err
	}
	r.localServices[service.ID] = service
	return nil
}

func (r *etcdRegistry) UnregisterService(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityService, id)
	if err := r.backend.deleteKey(ctx, key); err != nil {
		return err
	}
	delete(r.localServices, id)
	return nil
}

func (r *etcdRegistry) RegisterEndpoint(ctx context.Context, endpoint discovery.EndpointMeta) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityEndpoint, endpoint.ID)
	leaseID := r.etcdComp.LeaseID()
	if err := r.backend.putWithLease(ctx, key, endpoint, leaseID); err != nil {
		return err
	}
	r.localEndpoints[endpoint.ID] = endpoint
	return nil
}

func (r *etcdRegistry) UnregisterEndpoint(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityEndpoint, id)
	if err := r.backend.deleteKey(ctx, key); err != nil {
		return err
	}
	delete(r.localEndpoints, id)
	return nil
}

func (r *etcdRegistry) RegisterWorker(ctx context.Context, worker discovery.WorkerMeta) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityWorker, worker.ID)
	leaseID := r.etcdComp.LeaseID()
	if err := r.backend.putWithLease(ctx, key, worker, leaseID); err != nil {
		return err
	}
	r.localWorkers[worker.ID] = worker
	return nil
}

func (r *etcdRegistry) UnregisterWorker(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := entityKey(r.backend.options.Scope, entityWorker, id)
	if err := r.backend.deleteKey(ctx, key); err != nil {
		return err
	}
	delete(r.localWorkers, id)
	return nil
}

func (r *etcdRegistry) reRegisterAll(ctx context.Context, leaseID clientv3.LeaseID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log := runtime.RT.FrameLogger

	for id, meta := range r.localNodes {
		key := entityKey(r.backend.options.Scope, entityNode, id)
		if err := r.backend.putWithLease(ctx, key, meta, leaseID); err != nil {
			log.Error("EtcdRegistry", err, "failed to re-register node during reconnect")
		}
	}

	for id, meta := range r.localServices {
		key := entityKey(r.backend.options.Scope, entityService, id)
		if err := r.backend.putWithLease(ctx, key, meta, leaseID); err != nil {
			log.Error("EtcdRegistry", err, "failed to re-register service during reconnect")
		}
	}

	for id, meta := range r.localWorkers {
		key := entityKey(r.backend.options.Scope, entityWorker, id)
		if err := r.backend.putWithLease(ctx, key, meta, leaseID); err != nil {
			log.Error("EtcdRegistry", err, "failed to re-register worker during reconnect")
		}
	}

	for id, meta := range r.localEndpoints {
		key := entityKey(r.backend.options.Scope, entityEndpoint, id)
		if err := r.backend.putWithLease(ctx, key, meta, leaseID); err != nil {
			log.Error("EtcdRegistry", err, "failed to re-register endpoint during reconnect")
		}
	}
}

var _ discovery.Registry = (*etcdRegistry)(nil)
