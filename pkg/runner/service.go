package runner

import (
	"context"
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner/types"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility"
)

type BaseService[R types.Runner] struct {
	*BaseWorker[R]
	labels    utility.Labels
	listeners []*rpc.Listener
	lisnMu    sync.Mutex
}

func NewService[R types.Runner](name types.ServiceName, runner R, opts types.ServiceOptions) *BaseService[R] {
	w := NewWorker[R](types.WorkerName(name), runner, opts.WorkerOptions)

	s := &BaseService[R]{
		BaseWorker: w,
		labels:     opts.Labels,
	}

	if aware, ok := any(runner).(types.ServiceAware); ok {
		aware.SetService(s)
	} else if aware, ok := any(runner).(types.WorkerAware); ok {
		aware.SetWorker(s)
	}

	return s
}

func (s *BaseService[R]) Start(ctx context.Context) error {
	return s.BaseWorker.Start(ctx)
}

func (s *BaseService[R]) InstallListener(ctx context.Context, l *rpc.Listener) error {
	if err := l.Start(ctx); err != nil {
		return err
	}

	s.lisnMu.Lock()
	s.listeners = append(s.listeners, l)
	s.lisnMu.Unlock()

	stateCh := l.LifeCycle.Listen()
	go func() {
		for state := range stateCh {
			switch state {
			case rpc.ListenerStateReady, rpc.ListenerStateStopping:
				s.lisnMu.Lock()
				epMeta := discovery.NewEndpointMetaFromListener(l.GetMetaInfo())
				epMeta.TargetID = s.GetId()
				epMeta.TargetName = string(s.Name)
				epMeta.Weight = 100
				s.lisnMu.Unlock()
				reg := runtime.RT.GetDiscoveryRegistry()
				if reg != nil {
					reg.RegisterEndpoint(context.Background(), epMeta)
				}
			case rpc.ListenerStateStopped, rpc.ListenerStateError:
				reg := runtime.RT.GetDiscoveryRegistry()
				if reg != nil {
					reg.UnregisterEndpoint(context.Background(), l.Id())
				}
				l.LifeCycle.RemoveListen(stateCh)
				return
			}
		}
	}()

	return nil
}

func (s *BaseService[R]) stopListeners() {
	for _, l := range s.listeners {
		l.Stop()
	}
}

func (s *BaseService[R]) Stop() error {
	s.running.Store(false)

	if err := s.LifeCycle.SetState(types.WorkerStateStopping); err != nil {
		return err
	}

	s.stopListeners()

	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()

	if err := s.runner.Shutdown(); err != nil {
		return err
	}

	s.stopProviders()
	s.disconnectComponents()

	if err := s.LifeCycle.SetState(types.WorkerStateStopped); err != nil {
		return err
	}

	return nil
}

func (s *BaseService[R]) ListenLifeCycle() chan types.WorkerState {
	return s.LifeCycle.Listen()
}

func (s *BaseService[R]) GetMetadata() types.WorkerMetaData {
	meta := s.BaseWorker.GetMetadata()
	meta.Labels = s.labels
	return meta
}

func (s *BaseService[R]) Runner() R {
	return s.BaseWorker.Runner()
}
