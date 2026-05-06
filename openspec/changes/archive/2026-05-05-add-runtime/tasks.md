## 1. Package Setup

- [x] 1.1 Create `pkg/runtime/` package directory
- [x] 1.2 Create `pkg/runtime/runtime.go` with Runtime struct definition: startTime, root, FrameLogger, RpcLogger, components map, services map, workers map, and corresponding mutex fields

## 2. Constructor & Singleton

- [x] 2.1 Implement `NewRuntime()` — zero-arg constructor that captures `time.Now()`, `os.Getwd()`, creates `FrameLogger` (identify="framework") and `RpcLogger` (identify="rpc"), initializes empty maps
- [x] 2.2 Declare package-level convention singleton `var RT = NewRuntime()`
- [x] 2.3 Implement `StartTime() time.Time` and `Root() string` getter methods

## 3. Component Registry

- [x] 3.1 Implement `RegisterComponent(name string, c component.Component)` — store in map with RWMutex protection, do NOT call Start
- [x] 3.2 Implement `GetComponent(name string) (component.Component, bool)` — lookup in map with RWMutex protection

## 4. Worker/Service Registry

- [x] 4.1 Implement `InstallService(svc runner.Service)` — store in map keyed by `svc.GetMetadata().Id` with RWMutex protection
- [x] 4.2 Implement `InstallWorker(w runner.Worker)` — store in map keyed by `w.GetMetadata().Id` with RWMutex protection
- [x] 4.3 Implement `UninstallService(id string) error` — lookup, call Stop(), remove from map, return Stop error. No-op if not found
- [x] 4.4 Implement `UninstallWorker(id string) error` — lookup, call Stop(), remove from map, return Stop error. No-op if not found

## 5. Lifecycle Placeholders

- [x] 5.1 Implement `Startup(ctx context.Context) error` — return nil
- [x] 5.2 Implement `Shutdown() error` — return nil

## 6. Verification

- [x] 6.1 Verify compilation: `go build ./pkg/runtime/...`
- [x] 6.2 Verify imports from `pkg/component`, `pkg/runner`, `pkg/logger` resolve correctly
- [x] 6.3 Update `cmd/sora-test/main.go` to use `runtime.RT` as a smoke test
