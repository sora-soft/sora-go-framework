## MODIFIED Requirements

### Requirement: Worker lifecycle context propagation
Worker SHALL accept an external `context.Context` via `Start(ctx context.Context)` and derive a cancellable lifecycle context from it. The lifecycle context SHALL survive from `Start()` until `Stop()` is called. Upon construction via `NewWorker`, the worker SHALL automatically read `NodeId` from `runtime.RT.NodeId()` and populate `WorkerMetaData.NodeId`.

#### Scenario: Start creates lifecycle context from parent
- **WHEN** `Start(ctx)` is called with a valid parent context
- **THEN** the worker SHALL create an internal lifecycle context derived from the parent via `context.WithCancel(ctx)`

#### Scenario: Parent context cancellation propagates to worker
- **WHEN** the parent context passed to `Start()` is cancelled
- **THEN** the worker's lifecycle context SHALL also be cancelled

#### Scenario: Lifecycle context is used by Startup
- **WHEN** `Runner.Startup()` is called during `Start()`
- **THEN** it SHALL receive the worker's lifecycle context as its argument

#### Scenario: NewWorker populates NodeId from Runtime
- **WHEN** `NewWorker(name, runner, opts)` is called and `runtime.RT.NodeId()` returns a non-empty string
- **THEN** the created worker's `GetMetadata().NodeId` SHALL equal the Runtime's NodeId

#### Scenario: NewWorker with empty NodeId
- **WHEN** `NewWorker(name, runner, opts)` is called and `runtime.RT.NodeId()` returns an empty string
- **THEN** the created worker's `GetMetadata().NodeId` SHALL be an empty string
