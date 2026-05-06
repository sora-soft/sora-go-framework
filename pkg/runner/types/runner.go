package types

import (
	"context"

	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
)

type Runner interface {
	Startup(context.Context) error
	Shutdown() error
}

type Worker interface {
	GetId() string
	Start(ctx context.Context) error
	Stop() error
	Go(fn func(ctx context.Context))
	GetMetadata() WorkerMetaData
}

type Service interface {
	Worker
}

type WorkerRef interface {
	ConnectComponent(ctx context.Context, c component.Component) error
}

type ServiceRef interface {
	WorkerRef
	InstallListener(ctx context.Context, l *rpc.Listener) error
}

type WorkerRefAware interface {
	SetWorkerRef(WorkerRef)
}

type ServiceRefAware interface {
	SetServiceRef(ServiceRef)
}

type LifeCycleListener interface {
	ListenLifeCycle() chan WorkerState
}
