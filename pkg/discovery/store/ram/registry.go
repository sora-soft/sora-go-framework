package ram

import (
	"context"

	"github.com/sora-soft/sora-go-framework/pkg/discovery"
)

type ramRegistry struct {
	store *store
}

func (r *ramRegistry) RegisterNode(_ context.Context, node discovery.NodeMeta) error {
	r.store.mu.Lock()
	r.store.nodes[node.ID] = node
	r.store.mu.Unlock()
	r.store.nodeNotifier.push(r.store.snapshotNodes())
	return nil
}

func (r *ramRegistry) UnregisterNode(_ context.Context, id string) error {
	r.store.mu.Lock()
	delete(r.store.nodes, id)
	r.store.mu.Unlock()
	r.store.nodeNotifier.push(r.store.snapshotNodes())
	return nil
}

func (r *ramRegistry) RegisterService(_ context.Context, service discovery.ServiceMeta) error {
	r.store.mu.Lock()
	r.store.services[service.ID] = service
	r.store.mu.Unlock()
	r.store.serviceNotifier.push(r.store.snapshotServices())
	return nil
}

func (r *ramRegistry) UnregisterService(_ context.Context, id string) error {
	r.store.mu.Lock()
	delete(r.store.services, id)
	r.store.mu.Unlock()
	r.store.serviceNotifier.push(r.store.snapshotServices())
	return nil
}

func (r *ramRegistry) RegisterEndpoint(_ context.Context, endpoint discovery.EndpointMeta) error {
	r.store.mu.Lock()
	r.store.endpoints[endpoint.ID] = endpoint
	r.store.mu.Unlock()
	r.store.endpointNotifier.push(r.store.snapshotEndpoints())
	return nil
}

func (r *ramRegistry) UnregisterEndpoint(_ context.Context, id string) error {
	r.store.mu.Lock()
	delete(r.store.endpoints, id)
	r.store.mu.Unlock()
	r.store.endpointNotifier.push(r.store.snapshotEndpoints())
	return nil
}

func (r *ramRegistry) RegisterWorker(_ context.Context, worker discovery.WorkerMeta) error {
	r.store.mu.Lock()
	r.store.workers[worker.ID] = worker
	r.store.mu.Unlock()
	r.store.workerNotifier.push(r.store.snapshotWorkers())
	return nil
}

func (r *ramRegistry) UnregisterWorker(_ context.Context, id string) error {
	r.store.mu.Lock()
	delete(r.store.workers, id)
	r.store.mu.Unlock()
	r.store.workerNotifier.push(r.store.snapshotWorkers())
	return nil
}
