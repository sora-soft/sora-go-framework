## 1. Extract runner/types sub-package

- [x] 1.1 Create `pkg/runner/types/` directory
- [x] 1.2 Move `Runner`, `Worker`, `Service`, `WorkerRef`, `ServiceRef`, `WorkerRefAware`, `ServiceRefAware` interfaces from `pkg/runner/interface.go` to `pkg/runner/types/runner.go`
- [x] 1.3 Move `WorkerMetaData`, `WorkerState`, `WorkerState` constants from `pkg/runner/woker.go` to `pkg/runner/types/metadata.go`
- [x] 1.4 Move `WorkerOptions`, `ServiceOptions` structs to `pkg/runner/types/options.go`
- [x] 1.5 Update `pkg/runner/woker.go` imports to reference `types` sub-package
- [x] 1.6 Update `pkg/runner/service.go` imports to reference `types` sub-package
- [x] 1.7 Delete `pkg/runner/interface.go` (contents moved to types)

## 2. Migrate pkg/runtime to pkg/runner/runtime

- [x] 2.1 Create `pkg/runner/runtime/` directory
- [x] 2.2 Move `Runtime` struct and `RT` singleton from `pkg/runtime/runtime.go` to `pkg/runner/runtime/runtime.go`
- [x] 2.3 Update `Runtime` struct: change `services map[string]runner.Service` to `map[string]types.Service`, `workers map[string]runner.Worker` to `map[string]types.Worker`, `components map[string]component.Component` unchanged
- [x] 2.4 Add `nodeId string` field to `Runtime` struct
- [x] 2.5 Add `NodeId() string` and `SetNodeId(id string)` methods to `Runtime`
- [x] 2.6 Delete `pkg/runtime/` directory

## 3. Update existing runner code for new imports

- [x] 3.1 Update `pkg/runner/woker.go`: import `runner/types` and `runner/runtime`, update all type references (Runner → types.Runner, WorkerMetaData → types.WorkerMetaData, etc.)
- [x] 3.2 Update `pkg/runner/service.go`: import `runner/types`, update all type references
- [x] 3.3 Update `NewWorker` to read `nodeId` from `runtime.RT.NodeId()` and set it in `baseWorker.nodeId`
- [x] 3.4 Update `baseWorker.GetMetadata()` to return `baseWorker.nodeId` instead of empty string

## 4. Implement NodeRunner

- [x] 4.1 Create `pkg/runner/node.go` with `NodeOptions` struct (Alias *string, Version string), `NodeMetaData`, `NodeVersions`, `NodeRunData` structs
- [x] 4.2 Create `NodeRunner` struct with fields: options, listeners, svcRef, svc
- [x] 4.3 Implement `NodeRunner.Startup(ctx)`: iterate listeners, call svcRef.InstallListener; call runtime.RT.SetNodeId with svc.GetMetadata().Id
- [x] 4.4 Implement `NodeRunner.Shutdown()`: return nil
- [x] 4.5 Implement `NodeRunner.SetServiceRef(svcRef)`: store ref
- [x] 4.6 Implement `NodeRunner.SetService(svc)`: store service reference for reading ID
- [x] 4.7 Implement `NodeRunner.StateData() NodeMetaData`: collect Host, Pid, build from svc.GetMetadata()
- [x] 4.8 Implement `NodeRunner.RunData() NodeRunData`: aggregate from runtime.RT

## 5. Update external references

- [x] 5.1 Update `cmd/sora-test/main.go`: change `import runtime` to `import runner/runtime`, update all `runtime.RT` to `runtime.RT` via new import alias
- [x] 5.2 Update `cmd/sora-cli/main.go` if it references `pkg/runtime`
- [x] 5.3 Update any other files referencing `pkg/runtime` or `runner.Worker`/`runner.Service` directly

## 6. Verify

- [x] 6.1 Run `go build ./...` to verify compilation
- [x] 6.2 Run existing tests (if any) to verify no regressions
