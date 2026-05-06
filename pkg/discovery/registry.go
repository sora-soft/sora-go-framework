package discovery

import "context"

type Registry interface {
	RegisterNode(ctx context.Context, node NodeMeta) error
	UnregisterNode(ctx context.Context, id string) error
	RegisterService(ctx context.Context, service ServiceMeta) error
	UnregisterService(ctx context.Context, id string) error
	RegisterEndpoint(ctx context.Context, endpoint EndpointMeta) error
	UnregisterEndpoint(ctx context.Context, id string) error
	RegisterWorker(ctx context.Context, worker WorkerMeta) error
	UnregisterWorker(ctx context.Context, id string) error
}
