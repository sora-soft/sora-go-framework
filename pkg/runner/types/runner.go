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
	ConnectComponent(ctx context.Context, c component.Component) error
}

type Service interface {
	Worker
	InstallListener(ctx context.Context, l *rpc.Listener) error
}

type WorkerAware interface {
	SetWorker(Worker)
}

type ServiceAware interface {
	SetService(Service)
}

type LifeCycleListener interface {
	ListenLifeCycle() chan WorkerState
}
