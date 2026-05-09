package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/provider"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner/types"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility"
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
	if err := c.Start(ctx); err != nil {
		return err
	}
	b.compMu.Lock()
	b.components = append(b.components, c)
	b.compMu.Unlock()
	return nil
}

func (b *BaseWorker[R]) RegisterProvider(ctx context.Context, p provider.Provider) error {
	if err := p.Start(ctx); err != nil {
		return err
	}
	b.provMu.Lock()
	b.providers = append(b.providers, p)
	b.provMu.Unlock()
	return nil
}

func (b *BaseWorker[R]) disconnectComponents() {
	for _, c := range b.components {
		c.Stop()
	}
}

func (b *BaseWorker[R]) stopProviders() {
	for _, p := range b.providers {
		p.Stop()
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
