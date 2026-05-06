### Requirement: ConnectComponent immediately connects
`ConnectComponent(ctx, comp)` SHALL immediately call `comp.Start(ctx)`. On success, the component SHALL be recorded for lifecycle management. On failure, the component SHALL NOT be recorded and the error SHALL be returned.

#### Scenario: Successful component connection
- **WHEN** `ConnectComponent(ctx, comp)` is called and `comp.Start(ctx)` succeeds
- **THEN** the component SHALL be added to the internal components slice and nil error SHALL be returned

#### Scenario: Failed component connection
- **WHEN** `ConnectComponent(ctx, comp)` is called and `comp.Start(ctx)` fails
- **THEN** the component SHALL NOT be added to the internal components slice and the error SHALL be returned

### Requirement: ConnectComponent is concurrency-safe
The components slice SHALL be protected by a mutex, allowing concurrent calls to `ConnectComponent`.

#### Scenario: Concurrent ConnectComponent calls
- **WHEN** multiple goroutines call `ConnectComponent` simultaneously
- **THEN** all successful components SHALL be recorded without data race

### Requirement: ConnectComponent callable at any time
`ConnectComponent` SHALL be callable during `Startup()` or at any later time, not limited to the Start lifecycle phase.

#### Scenario: ConnectComponent called during Startup
- **WHEN** a Runner calls `ref.ConnectComponent(ctx, comp)` inside its `Startup()` method
- **THEN** the component SHALL be connected immediately and recorded

#### Scenario: ConnectComponent called after Start completes
- **WHEN** `ConnectComponent(ctx, comp)` is called after `Start()` has returned successfully
- **THEN** the component SHALL be connected immediately and recorded

### Requirement: disconnectComponents on Stop
`baseWorker.Stop()` SHALL call `disconnectComponents()` which iterates all recorded components and calls `comp.Stop()` on each.

#### Scenario: All components disconnected on Stop
- **WHEN** `Stop()` is called and there are 3 recorded components
- **THEN** `Stop()` SHALL be called on all 3 components in the order they were registered

### Requirement: disconnectComponents on Startup failure
When `Runner.Startup()` returns an error, `baseWorker.Start()` SHALL call `disconnectComponents()` to clean up all successfully connected components before setting Error state.

#### Scenario: Startup failure cleans up partial connections
- **WHEN** `Startup()` returns an error after 2 components were successfully registered
- **THEN** both components SHALL be disconnected (Stop called) and state SHALL be set to Error
