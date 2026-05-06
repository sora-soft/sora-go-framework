package etcd

import (
	"context"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
)

type etcdRegistry struct {
	backend *EtcdBackend
}

func (r *etcdRegistry) RegisterNode(ctx context.Context, node discovery.NodeMeta) error {
	key := entityKey(r.backend.options.Prefix, entityNode, node.ID)
	return r.backend.putWithLease(ctx, key, node)
}

func (r *etcdRegistry) UnregisterNode(ctx context.Context, id string) error {
	key := entityKey(r.backend.options.Prefix, entityNode, id)
	return r.backend.deleteKey(ctx, key)
}

func (r *etcdRegistry) RegisterService(ctx context.Context, service discovery.ServiceMeta) error {
	key := entityKey(r.backend.options.Prefix, entityService, service.ID)
	return r.backend.putWithLease(ctx, key, service)
}

func (r *etcdRegistry) UnregisterService(ctx context.Context, id string) error {
	key := entityKey(r.backend.options.Prefix, entityService, id)
	return r.backend.deleteKey(ctx, key)
}

func (r *etcdRegistry) RegisterEndpoint(ctx context.Context, endpoint discovery.EndpointMeta) error {
	key := entityKey(r.backend.options.Prefix, entityEndpoint, endpoint.ID)
	return r.backend.putWithLease(ctx, key, endpoint)
}

func (r *etcdRegistry) UnregisterEndpoint(ctx context.Context, id string) error {
	key := entityKey(r.backend.options.Prefix, entityEndpoint, id)
	return r.backend.deleteKey(ctx, key)
}

func (r *etcdRegistry) RegisterWorker(ctx context.Context, worker discovery.WorkerMeta) error {
	key := entityKey(r.backend.options.Prefix, entityWorker, worker.ID)
	return r.backend.putWithLease(ctx, key, worker)
}

func (r *etcdRegistry) UnregisterWorker(ctx context.Context, id string) error {
	key := entityKey(r.backend.options.Prefix, entityWorker, id)
	return r.backend.deleteKey(ctx, key)
}

var _ discovery.Registry = (*etcdRegistry)(nil)
