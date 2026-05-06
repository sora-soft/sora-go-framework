package ram

import (
	"context"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
)

type ramDiscovery struct {
	store *store
}

func (d *ramDiscovery) GetNode(_ context.Context, id string) (*discovery.NodeMeta, error) {
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	if n, ok := d.store.nodes[id]; ok {
		return &n, nil
	}
	return nil, nil
}

func (d *ramDiscovery) ListNodes(_ context.Context) ([]discovery.NodeMeta, error) {
	return d.store.snapshotNodes(), nil
}

func (d *ramDiscovery) WatchNodes(ctx context.Context) <-chan []discovery.NodeMeta {
	return watchEntities(ctx, &d.store.nodeNotifier, d.store.snapshotNodes)
}

func (d *ramDiscovery) GetService(_ context.Context, id string) (*discovery.ServiceMeta, error) {
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	if svc, ok := d.store.services[id]; ok {
		return &svc, nil
	}
	return nil, nil
}

func (d *ramDiscovery) ListServices(_ context.Context) ([]discovery.ServiceMeta, error) {
	return d.store.snapshotServices(), nil
}

func (d *ramDiscovery) ListServicesByName(_ context.Context, name string) ([]discovery.ServiceMeta, error) {
	all := d.store.snapshotServices()
	result := make([]discovery.ServiceMeta, 0)
	for _, svc := range all {
		if svc.Name == name {
			result = append(result, svc)
		}
	}
	return result, nil
}

func (d *ramDiscovery) WatchServices(ctx context.Context) <-chan []discovery.ServiceMeta {
	return watchEntities(ctx, &d.store.serviceNotifier, d.store.snapshotServices)
}

func (d *ramDiscovery) GetWorker(_ context.Context, id string) (*discovery.WorkerMeta, error) {
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	if w, ok := d.store.workers[id]; ok {
		return &w, nil
	}
	return nil, nil
}

func (d *ramDiscovery) ListWorkers(_ context.Context) ([]discovery.WorkerMeta, error) {
	return d.store.snapshotWorkers(), nil
}

func (d *ramDiscovery) ListWorkersByName(_ context.Context, name string) ([]discovery.WorkerMeta, error) {
	all := d.store.snapshotWorkers()
	result := make([]discovery.WorkerMeta, 0)
	for _, w := range all {
		if w.Name == name {
			result = append(result, w)
		}
	}
	return result, nil
}

func (d *ramDiscovery) WatchWorkers(ctx context.Context) <-chan []discovery.WorkerMeta {
	return watchEntities(ctx, &d.store.workerNotifier, d.store.snapshotWorkers)
}

func (d *ramDiscovery) GetEndpoint(_ context.Context, id string) (*discovery.EndpointMeta, error) {
	d.store.mu.RLock()
	defer d.store.mu.RUnlock()
	if ep, ok := d.store.endpoints[id]; ok {
		return &ep, nil
	}
	return nil, nil
}

func (d *ramDiscovery) ListEndpoints(_ context.Context) ([]discovery.EndpointMeta, error) {
	return d.store.snapshotEndpoints(), nil
}

func (d *ramDiscovery) ListEndpointsByService(_ context.Context, serviceID string) ([]discovery.EndpointMeta, error) {
	all := d.store.snapshotEndpoints()
	result := make([]discovery.EndpointMeta, 0)
	for _, ep := range all {
		if ep.TargetID == serviceID {
			result = append(result, ep)
		}
	}
	return result, nil
}

func (d *ramDiscovery) WatchEndpoints(ctx context.Context) <-chan []discovery.EndpointMeta {
	return watchEntities(ctx, &d.store.endpointNotifier, d.store.snapshotEndpoints)
}
