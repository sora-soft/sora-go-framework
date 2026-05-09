package runtime

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/logger"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner/types"
)

type Runtime struct {
	startTime   time.Time
	root        string
	FrameLogger *logger.Logger
	RpcLogger   *logger.Logger

	components map[component.ComponentName]component.Component
	compMu     sync.RWMutex

	services map[string]types.Service
	svcMu    sync.RWMutex

	workers  map[string]types.Worker
	workerMu sync.RWMutex

	nodeMu  sync.RWMutex
	node    types.Service
	backend discovery.Backend
}

func NewRuntime() *Runtime {
	return &Runtime{
		startTime:   time.Now(),
		root:        mustGetwd(),
		FrameLogger: logger.NewLogger("framework"),
		RpcLogger:   logger.NewLogger("rpc"),
		components:  make(map[component.ComponentName]component.Component),
		services:    make(map[string]types.Service),
		workers:     make(map[string]types.Worker),
	}
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

var RT = NewRuntime()

func (r *Runtime) StartTime() time.Time {
	return r.startTime
}

func (r *Runtime) Root() string {
	return r.root
}

func (r *Runtime) NodeId() string {
	r.nodeMu.RLock()
	defer r.nodeMu.RUnlock()
	if r.node == nil {
		return ""
	}
	return r.node.GetId()
}

func (r *Runtime) RegisterComponent(name component.ComponentName, c component.Component) {
	r.compMu.Lock()
	r.components[name] = c
	r.compMu.Unlock()
}

func (r *Runtime) GetComponent(name component.ComponentName) (component.Component, bool) {
	r.compMu.RLock()
	c, ok := r.components[name]
	r.compMu.RUnlock()
	return c, ok
}

func GetComponentOf[T component.ComponentImpl](name component.ComponentName) (*component.BaseComponent[T], error) {
	c, ok := RT.GetComponent(name)
	if !ok {
		return nil, fmt.Errorf("component not found: %s", name)
	}
	bc, ok := c.(*component.BaseComponent[T])
	if !ok {
		return nil, fmt.Errorf("component type mismatch: %s", name)
	}
	return bc, nil
}

func (r *Runtime) GetAllComponents() []component.Component {
	r.compMu.RLock()
	result := make([]component.Component, 0, len(r.components))
	for _, c := range r.components {
		result = append(result, c)
	}
	r.compMu.RUnlock()
	return result
}

func (r *Runtime) InstallService(ctx context.Context, svc types.Service) error {
	r.svcMu.Lock()
	r.services[svc.GetId()] = svc
	r.svcMu.Unlock()

	if listener, ok := svc.(types.LifeCycleListener); ok {
		ch := listener.ListenLifeCycle()
		go func() {
			for state := range ch {
				switch state {
				case types.WorkerStateReady, types.WorkerStateInit, types.WorkerStatePending, types.WorkerStateStopping, types.WorkerStateError:
					reg := r.GetDiscoveryRegistry()
					if reg != nil {
						meta := svc.GetMetadata()
						svcMeta := discovery.ServiceMeta{
							Name:      string(meta.Name),
							ID:        meta.Id,
							Alias:     meta.Alias,
							State:     int(meta.State),
							NodeID:    r.NodeId(),
							StartTime: meta.StartTime,
							Labels:    meta.Labels,
						}
						reg.RegisterService(context.Background(), svcMeta)
					}
				case types.WorkerStateStopped:
					reg := r.GetDiscoveryRegistry()
					if reg != nil {
						reg.UnregisterService(context.Background(), svc.GetId())
					}
				}
			}
		}()
	}

	return svc.Start(ctx)
}

func (r *Runtime) InstallWorker(w types.Worker) {
	r.workerMu.Lock()
	r.workers[w.GetId()] = w
	r.workerMu.Unlock()

	if listener, ok := w.(types.LifeCycleListener); ok {
		ch := listener.ListenLifeCycle()
		go func() {
			for state := range ch {
				switch state {
				case types.WorkerStateReady, types.WorkerStateInit, types.WorkerStatePending, types.WorkerStateStopping, types.WorkerStateError:
					reg := r.GetDiscoveryRegistry()
					if reg != nil {
						meta := w.GetMetadata()
						workerMeta := discovery.WorkerMeta{
							Name:      string(meta.Name),
							ID:        meta.Id,
							Alias:     meta.Alias,
							State:     int(meta.State),
							NodeID:    r.NodeId(),
							StartTime: meta.StartTime,
						}
						reg.RegisterWorker(context.Background(), workerMeta)
					}
				case types.WorkerStateStopped:
					reg := r.GetDiscoveryRegistry()
					if reg != nil {
						reg.UnregisterWorker(context.Background(), w.GetId())
					}
				}
			}
		}()
	}
}

func (r *Runtime) UninstallService(id string) error {
	r.svcMu.Lock()
	svc, ok := r.services[id]
	if !ok {
		r.svcMu.Unlock()
		return nil
	}
	delete(r.services, id)
	r.svcMu.Unlock()

	return svc.Stop()
}

func (r *Runtime) UninstallWorker(id string) error {
	r.workerMu.Lock()
	w, ok := r.workers[id]
	if !ok {
		r.workerMu.Unlock()
		return nil
	}
	delete(r.workers, id)
	r.workerMu.Unlock()

	return w.Stop()
}

func (r *Runtime) GetAllServices() []types.Service {
	r.svcMu.RLock()
	result := make([]types.Service, 0, len(r.services))
	for _, svc := range r.services {
		result = append(result, svc)
	}
	r.svcMu.RUnlock()
	return result
}

func (r *Runtime) GetAllWorkers() []types.Worker {
	r.workerMu.RLock()
	result := make([]types.Worker, 0, len(r.workers))
	for _, w := range r.workers {
		result = append(result, w)
	}
	r.workerMu.RUnlock()
	return result
}

func (r *Runtime) GetNode() types.Service {
	r.nodeMu.RLock()
	defer r.nodeMu.RUnlock()
	return r.node
}

func (r *Runtime) GetBackend() discovery.Backend {
	r.nodeMu.RLock()
	defer r.nodeMu.RUnlock()
	return r.backend
}

func (r *Runtime) GetDiscovery() discovery.Discovery {
	r.nodeMu.RLock()
	defer r.nodeMu.RUnlock()
	if r.backend == nil {
		return nil
	}
	return r.backend.Discovery()
}

func (r *Runtime) GetDiscoveryRegistry() discovery.Registry {
	r.nodeMu.RLock()
	defer r.nodeMu.RUnlock()
	if r.backend == nil {
		return nil
	}
	return r.backend.Registry()
}

func (r *Runtime) Startup(ctx context.Context, node types.Service, backend discovery.Backend) error {
	if err := backend.Connect(ctx); err != nil {
		return err
	}

	r.nodeMu.Lock()
	r.node = node
	r.backend = backend
	r.nodeMu.Unlock()

	return r.InstallService(ctx, node)
}

func (r *Runtime) Shutdown() error {
	var firstErr error
	collectErr := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	nodeId := r.NodeId()

	r.svcMu.RLock()
	svcIDs := make([]string, 0, len(r.services))
	for id := range r.services {
		if id != nodeId {
			svcIDs = append(svcIDs, id)
		}
	}
	r.svcMu.RUnlock()

	r.workerMu.RLock()
	workerIDs := make([]string, 0, len(r.workers))
	for id := range r.workers {
		workerIDs = append(workerIDs, id)
	}
	r.workerMu.RUnlock()

	var wg sync.WaitGroup
	for _, id := range svcIDs {
		wg.Add(1)
		go func(sid string) {
			defer wg.Done()
			collectErr(r.UninstallService(sid))
		}(id)
	}
	for _, id := range workerIDs {
		wg.Add(1)
		go func(wid string) {
			defer wg.Done()
			collectErr(r.UninstallWorker(wid))
		}(id)
	}
	wg.Wait()

	if nodeId != "" {
		collectErr(r.UninstallService(nodeId))
	}

	r.nodeMu.Lock()
	b := r.backend
	r.node = nil
	r.backend = nil
	r.nodeMu.Unlock()

	if b != nil {
		collectErr(b.Disconnect())
	}

	return firstErr
}
