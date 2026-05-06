## 1. Interface Definitions

- [x] 1.1 Add `Worker`, `Service` public interfaces and `WorkerRef`, `ServiceRef`, `WorkerRefAware`, `ServiceRefAware` interfaces to `pkg/runner/interface.go`
- [x] 1.2 Update `WorkerMetaData` struct: add `Labels utility.Labels` and `Listeners []rpc.ListenerMetaInfo` fields with `omitempty` JSON tags
- [x] 1.3 Verify compilation: all new interfaces compile, existing `Runner` interface unchanged

## 2. baseWorker Refactoring

- [x] 2.1 Add `components []component.Component` and `compMu sync.Mutex` fields to `baseWorker`
- [x] 2.2 Implement `ConnectComponent(ctx, comp)` method on `baseWorker`: call `comp.Start(ctx)`, record on success, return error on failure
- [x] 2.3 Implement `disconnectComponents()` method on `baseWorker`: iterate components, call `Stop()` on each
- [x] 2.4 Update `baseWorker.Start()`: on `Runner.Startup()` failure, call `disconnectComponents()` before setting Error state
- [x] 2.5 Update `baseWorker.Stop()`: call `disconnectComponents()` after `Runner.Shutdown()`
- [x] 2.6 Update `NewWorker`: inject `WorkerRef` via `WorkerRefAware` interface assertion; return `Worker` interface type instead of `*baseWorker`
- [x] 2.7 Update `baseWorker.GetMetadata()`: return `WorkerMetaData` with Labels and Listeners as zero values

## 3. baseService Implementation

- [x] 3.1 Rename `BaseService` to `baseService` (unexported), add `labels utility.Labels`, `listeners []*rpc.Listener`, `lisnMu sync.Mutex` fields
- [x] 3.2 Implement `InstallListener(ctx, ln)` method on `baseService`: call `ln.Start(ctx)`, record on success, return error on failure
- [x] 3.3 Implement `stopListeners()` method on `baseService`: iterate listeners, call `Stop()` on each
- [x] 3.4 Implement `baseService.Stop()` override: set Stopping → set running false → stopListeners → cancel → wg.Wait → Runner.Shutdown → disconnectComponents → set Stopped
- [x] 3.5 Implement `baseService.GetMetadata()` override: populate Labels and Listeners from service state
- [x] 3.6 Implement `NewService` constructor: create baseService wrapping NewWorker internals, inject ServiceRef/WorkerRef via interface assertion, return `Service` interface type

## 4. Verification

- [x] 4.1 Verify `Service` interface is assignable to `Worker` (compile-time check)
- [x] 4.2 Verify external code cannot reference `baseWorker` or `baseService` (unexported check)
- [x] 4.3 Run existing tests and verify no regressions
- [x] 4.4 Update `cmd/sora-test/main.go` if it references `baseWorker` or `BaseService` directly
