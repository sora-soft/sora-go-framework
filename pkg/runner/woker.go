package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sora-soft/sora-go-framework/pkg/component"
	"github.com/sora-soft/sora-go-framework/pkg/logger"
	"github.com/sora-soft/sora-go-framework/pkg/rpc/provider"
	"github.com/sora-soft/sora-go-framework/pkg/runner/types"
	"github.com/sora-soft/sora-go-framework/pkg/runtime"
	"github.com/sora-soft/sora-go-framework/pkg/utility"
)

type BaseWorker[R types.Runner] struct {
	runner    R
	Name      types.WorkerName
	Id        string
	StartTime int64
	options   types.WorkerOptions

	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool

	LifeCycle  *utility.LifeCycle[types.WorkerState]
	components []component.Component
	compMu     sync.Mutex
	providers  []provider.Provider
	provMu     sync.Mutex
}

func (b *BaseWorker[R]) ConnectComponent(ctx context.Context, c component.Component) error {
	meta := c.GetMetaInfo()
	runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "connect-component", "id": b.Id, "name": b.Name, "component": meta.Name, "version": meta.Version})
	if err := c.Start(ctx); err != nil {
		return err
	}
	b.compMu.Lock()
	b.components = append(b.components, c)
	b.compMu.Unlock()
	runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "component-connected", "id": b.Id, "name": b.Name, "component": meta.Name, "version": meta.Version})
	return nil
}

func (b *BaseWorker[R]) RegisterProvider(ctx context.Context, p provider.Provider) error {
	runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "register-provider", "id": b.Id, "name": b.Name})
	if err := p.Start(ctx); err != nil {
		return err
	}
	b.provMu.Lock()
	b.providers = append(b.providers, p)
	b.provMu.Unlock()
	runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "provider-started", "id": b.Id, "name": b.Name})
	return nil
}

func (b *BaseWorker[R]) disconnectComponents() {
	for _, c := range b.components {
		meta := c.GetMetaInfo()
		runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "disconnect-component", "id": b.Id, "name": b.Name, "component": meta.Name})
		c.Stop()
		runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "component-disconnected", "id": b.Id, "name": b.Name, "component": meta.Name})
	}
}

func (b *BaseWorker[R]) stopProviders() {
	for _, p := range b.providers {
		runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "unregister-provider", "id": b.Id, "name": b.Name})
		p.Stop()
		runtime.RT.FrameLogger.Info("runtime", map[string]any{"event": "provider-unregistered", "id": b.Id, "name": b.Name})
	}
}

func NewWorker[R types.Runner](name types.WorkerName, runner R, options types.WorkerOptions) *BaseWorker[R] {
	w := &BaseWorker[R]{
		Name:      name,
		Id:        uuid.New().String(),
		StartTime: time.Now().Unix(),
		LifeCycle: utility.NewLifeCycle(types.WorkerStateInit, false),
		runner:    runner,
		options:   options,
	}

	if aware, ok := any(runner).(types.WorkerAware); ok {
		aware.SetWorker(w)
	}

	return w
}

func (b *BaseWorker[R]) Start(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			runtime.RT.FrameLogger.Error("runtime", err, map[string]any{"event": "worker-on-error", "error": logger.ErrorMessage(err), "name": b.Name, "id": b.Id})
			b.disconnectComponents()
			b.LifeCycle.SetStateWithError(types.WorkerStateError, err)
		}
	}()

	b.ctx, b.cancel = context.WithCancel(ctx)

	if err := b.LifeCycle.SetState(types.WorkerStatePending); err != nil {
		return err
	}

	b.running.Store(true)

	errCh := make(chan error, 1)
	go func() {
		errCh <- b.runner.Startup(b.ctx)
	}()

	if err := <-errCh; err != nil {
		return err
	}

	if err := b.LifeCycle.SetState(types.WorkerStateReady); err != nil {
		return err
	}
	return nil
}

func (b *BaseWorker[R]) Stop() error {
	b.running.Store(false)

	if err := b.LifeCycle.SetState(types.WorkerStateStopping); err != nil {
		return err
	}

	if b.cancel != nil {
		b.cancel()
	}
	b.wg.Wait()

	if err := b.runner.Shutdown(); err != nil {
		return err
	}

	b.stopProviders()
	b.disconnectComponents()

	if err := b.LifeCycle.SetState(types.WorkerStateStopped); err != nil {
		return err
	}

	return nil
}

func (b *BaseWorker[R]) Go(fn func(ctx context.Context)) {
	if !b.running.Load() {
		return
	}

	b.wg.Go(func() {
		fn(b.ctx)
	})
}

func (b *BaseWorker[R]) GetId() string {
	return b.Id
}

func (b *BaseWorker[R]) GetMetadata() types.WorkerMetaData {
	return types.WorkerMetaData{
		Name:      b.Name,
		Alias:     b.options.Alias,
		State:     b.LifeCycle.GetState(),
		Id:        b.Id,
		StartTime: b.StartTime,
	}
}

func (b *BaseWorker[R]) Runner() R {
	return b.runner
}
