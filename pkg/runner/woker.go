package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner/types"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility"
)

type baseWorker struct {
	types.Runner
	Name      string
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
}

func (b *baseWorker) ConnectComponent(ctx context.Context, c component.Component) error {
	if err := c.Start(ctx); err != nil {
		return err
	}
	b.compMu.Lock()
	b.components = append(b.components, c)
	b.compMu.Unlock()
	return nil
}

func (b *baseWorker) disconnectComponents() {
	for _, c := range b.components {
		c.Stop()
	}
}

func NewWorker(name string, runner types.Runner, options types.WorkerOptions) types.Worker {
	w := &baseWorker{
		Name:      name,
		Id:        uuid.New().String(),
		StartTime: time.Now().Unix(),
		LifeCycle: utility.NewLifeCycle(types.WorkerStateInit, false),
		Runner:    runner,
		options:   options,
	}

	if aware, ok := runner.(types.WorkerAware); ok {
		aware.SetWorker(w)
	}

	return w
}

func (b *baseWorker) Start(ctx context.Context) (err error) {
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
		errCh <- b.Runner.Startup(b.ctx)
	}()

	if err := <-errCh; err != nil {
		return err
	}

	if err := b.LifeCycle.SetState(types.WorkerStateReady); err != nil {
		return err
	}
	return nil
}

func (b *baseWorker) Stop() error {
	b.running.Store(false)

	if err := b.LifeCycle.SetState(types.WorkerStateStopping); err != nil {
		return err
	}

	if b.cancel != nil {
		b.cancel()
	}
	b.wg.Wait()

	if err := b.Runner.Shutdown(); err != nil {
		return err
	}

	b.disconnectComponents()

	if err := b.LifeCycle.SetState(types.WorkerStateStopped); err != nil {
		return err
	}

	return nil
}

func (b *baseWorker) Go(fn func(ctx context.Context)) {
	if !b.running.Load() {
		return
	}

	b.wg.Go(func() {
		fn(b.ctx)
	})
}

func (b *baseWorker) GetId() string {
	return b.Id
}

func (b *baseWorker) GetMetadata() types.WorkerMetaData {
	return types.WorkerMetaData{
		Name:      b.Name,
		Alias:     b.options.Alias,
		State:     b.LifeCycle.GetState(),
		Id:        b.Id,
		StartTime: b.StartTime,
	}
}
