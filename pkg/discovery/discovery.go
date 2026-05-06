package discovery

import "context"

type Discovery interface {
	GetNode(ctx context.Context, id string) (*NodeMeta, error)
	ListNodes(ctx context.Context) ([]NodeMeta, error)
	WatchNodes(ctx context.Context) <-chan []NodeMeta

	GetService(ctx context.Context, id string) (*ServiceMeta, error)
	ListServices(ctx context.Context) ([]ServiceMeta, error)
	ListServicesByName(ctx context.Context, name string) ([]ServiceMeta, error)
	WatchServices(ctx context.Context) <-chan []ServiceMeta

	GetWorker(ctx context.Context, id string) (*WorkerMeta, error)
	ListWorkers(ctx context.Context) ([]WorkerMeta, error)
	ListWorkersByName(ctx context.Context, name string) ([]WorkerMeta, error)
	WatchWorkers(ctx context.Context) <-chan []WorkerMeta

	GetEndpoint(ctx context.Context, id string) (*EndpointMeta, error)
	ListEndpoints(ctx context.Context) ([]EndpointMeta, error)
	ListEndpointsByService(ctx context.Context, serviceID string) ([]EndpointMeta, error)
	WatchEndpoints(ctx context.Context) <-chan []EndpointMeta
}
