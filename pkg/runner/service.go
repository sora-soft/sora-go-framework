package runner

import (
	"context"
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner/types"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility"
)

type baseService struct {
	*baseWorker
	labels    utility.Labels
	listeners []*rpc.Listener
	lisnMu    sync.Mutex
}

func NewService(name string, runner types.Runner, opts types.ServiceOptions) types.Service {
	w := NewWorker(name, runner, opts.WorkerOptions).(*baseWorker)

	s := &baseService{
		baseWorker: w,
		labels:     opts.Labels,
	}

	if aware, ok := runner.(types.ServiceRefAware); ok {
		aware.SetServiceRef(s)
	} else if aware, ok := runner.(types.WorkerRefAware); ok {
		aware.SetWorkerRef(s)
	}

	return s
}

func (s *baseService) Start(ctx context.Context) error {
	return s.baseWorker.Start(ctx)
}

func (s *baseService) InstallListener(ctx context.Context, l *rpc.Listener) error {
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
				epMeta.TargetName = s.Name
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

func (s *baseService) stopListeners() {
	for _, l := range s.listeners {
		l.Stop()
	}
}

func (s *baseService) Stop() error {
	s.running.Store(false)

	if err := s.LifeCycle.SetState(types.WorkerStateStopping); err != nil {
		return err
	}

	s.stopListeners()

	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()

	if err := s.Runner.Shutdown(); err != nil {
		return err
	}

	s.disconnectComponents()

	if err := s.LifeCycle.SetState(types.WorkerStateStopped); err != nil {
		return err
	}

	return nil
}

func (s *baseService) ListenLifeCycle() chan types.WorkerState {
	return s.LifeCycle.Listen()
}

func (s *baseService) GetMetadata() types.WorkerMetaData {
	meta := s.baseWorker.GetMetadata()
	meta.Labels = s.labels
	return meta
}
