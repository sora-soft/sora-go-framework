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

### Requirement: Worker task tracking via Go
Worker SHALL provide a `Go(fn func(ctx context.Context))` method that spawns tracked goroutines using the lifecycle context. All spawned tasks SHALL share the same lifecycle context.

#### Scenario: Go spawns task with lifecycle context
- **WHEN** `Go(fn)` is called while the worker is running (after Start, before Stop)
- **THEN** `fn` SHALL be executed in a new goroutine with the worker's lifecycle context as argument, and tracked via WaitGroup

#### Scenario: Go before Start is silently discarded
- **WHEN** `Go(fn)` is called before `Start()` has been called
- **THEN** `fn` SHALL NOT be executed and the call SHALL return silently

#### Scenario: Go after Stop is silently discarded
- **WHEN** `Go(fn)` is called after `Stop()` has been called
- **THEN** `fn` SHALL NOT be executed and the call SHALL return silently

### Requirement: Graceful stop with context cancellation
`Stop()` SHALL cancel the lifecycle context to signal all running tasks, wait for them to complete via WaitGroup, then invoke `Runner.Shutdown()`.

#### Scenario: Stop cancels context and waits for tasks
- **WHEN** `Stop()` is called while tasks are running
- **THEN** the worker SHALL (1) set running to false, (2) cancel the lifecycle context, (3) wait for all tracked goroutines to complete, (4) set state to Stopping, (5) call `Runner.Shutdown()`, (6) set state to Stopped

#### Scenario: Stop with no running tasks
- **WHEN** `Stop()` is called and no tasks are running via `Go()`
- **THEN** the worker SHALL proceed through the full stop sequence without blocking on WaitGroup

#### Scenario: Stop respects task graceful exit
- **WHEN** a task spawned via `Go()` is performing work and the lifecycle context is cancelled
- **THEN** the task SHALL be able to detect `ctx.Done()` and exit gracefully before `Stop()` proceeds to `Runner.Shutdown()`

### Requirement: Worker state machine transitions
Worker SHALL follow a specific state transition sequence during Start and Stop, using the existing WorkerState enum without modification. On Start failure, Worker SHALL disconnect all successfully connected components before setting Error state.

#### Scenario: Successful start state transitions
- **WHEN** `Start()` completes successfully
- **THEN** the state SHALL transition from Init → Pending → Ready

#### Scenario: Successful stop state transitions
- **WHEN** `Stop()` completes successfully
- **THEN** the state SHALL transition to Stopping → Stopped

#### Scenario: Start failure sets error state and cleans up components
- **WHEN** `Start()` encounters an error (either from state transition or from `Runner.Startup()`)
- **THEN** the state SHALL be set to Error with the associated error, and all successfully connected components SHALL be disconnected

### Requirement: Service registration on startup
When a ServiceRunner starts, it SHALL register itself and its endpoints with the discovery Registry if one is configured. Registration SHALL occur after listeners are started and the service state is Ready.

#### Scenario: Service registers with discovery on start
- **WHEN** a ServiceRunner starts with a discovery Registry configured
- **THEN** it SHALL call `Registry.RegisterService(ctx, serviceMeta)` with its metadata, followed by `Registry.RegisterEndpoint(ctx, endpointMeta)` for each active listener, where each endpoint's `TargetID` is set to the service ID and `TargetName` is set to the service name

#### Scenario: Service starts without discovery Registry
- **WHEN** a ServiceRunner starts without a discovery Registry configured
- **THEN** it SHALL proceed normally without calling any Registry methods

### Requirement: Service unregistration on shutdown
When a ServiceRunner stops, it SHALL unregister its endpoints and itself from the discovery Registry if one is configured. Unregistration SHALL occur in reverse order: endpoints first, then the service.

#### Scenario: Service unregisters from discovery on stop
- **WHEN** a ServiceRunner stops with a discovery Registry configured
- **THEN** it SHALL call `Registry.UnregisterEndpoint(ctx, endpointId)` for each registered endpoint, then `Registry.UnregisterService(ctx, serviceId)`

#### Scenario: Service stops without discovery Registry
- **WHEN** a ServiceRunner stops without a discovery Registry configured
- **THEN** it SHALL proceed with the normal shutdown sequence without calling any Registry methods

### Requirement: Service Stop overrides Worker Stop
`baseService.Stop()` SHALL stop all installed listeners before executing the Worker stop sequence (cancel context, wait for goroutines, Runner.Shutdown, disconnect components).

#### Scenario: Service Stop stops listeners then delegates to worker cleanup
- **WHEN** `Stop()` is called on a Service
- **THEN** the sequence SHALL be: set state Stopping → set running false → stopListeners → cancel context → wait for goroutines → Runner.Shutdown → disconnectComponents → set state Stopped
