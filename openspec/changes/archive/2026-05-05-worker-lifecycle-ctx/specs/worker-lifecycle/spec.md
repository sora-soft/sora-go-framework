## ADDED Requirements

### Requirement: Worker lifecycle context propagation
Worker SHALL accept an external `context.Context` via `Start(ctx context.Context)` and derive a cancellable lifecycle context from it. The lifecycle context SHALL survive from `Start()` until `Stop()` is called.

#### Scenario: Start creates lifecycle context from parent
- **WHEN** `Start(ctx)` is called with a valid parent context
- **THEN** the worker SHALL create an internal lifecycle context derived from the parent via `context.WithCancel(ctx)`

#### Scenario: Parent context cancellation propagates to worker
- **WHEN** the parent context passed to `Start()` is cancelled
- **THEN** the worker's lifecycle context SHALL also be cancelled

#### Scenario: Lifecycle context is used by Startup
- **WHEN** `Runner.Startup()` is called during `Start()`
- **THEN** it SHALL receive the worker's lifecycle context as its argument

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
Worker SHALL follow a specific state transition sequence during Start and Stop, using the existing WorkerState enum without modification.

#### Scenario: Successful start state transitions
- **WHEN** `Start()` completes successfully
- **THEN** the state SHALL transition from Init → Pending → Ready

#### Scenario: Successful stop state transitions
- **WHEN** `Stop()` completes successfully
- **THEN** the state SHALL transition to Stopping → Stopped

#### Scenario: Start failure sets error state
- **WHEN** `Start()` encounters an error (either from state transition or from `Runner.Startup()`)
- **THEN** the state SHALL be set to Error with the associated error
