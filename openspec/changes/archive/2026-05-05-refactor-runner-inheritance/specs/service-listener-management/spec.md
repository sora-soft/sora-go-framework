## ADDED Requirements

### Requirement: InstallListener immediately starts
`InstallListener(ctx, ln)` SHALL immediately call `ln.Start(ctx)`. On success, the listener SHALL be recorded for lifecycle management. On failure, the listener SHALL NOT be recorded and the error SHALL be returned.

#### Scenario: Successful listener installation
- **WHEN** `InstallListener(ctx, ln)` is called and `ln.Start(ctx)` succeeds
- **THEN** the listener SHALL be added to the internal listeners slice and nil error SHALL be returned

#### Scenario: Failed listener installation
- **WHEN** `InstallListener(ctx, ln)` is called and `ln.Start(ctx)` fails
- **THEN** the listener SHALL NOT be added to the internal listeners slice and the error SHALL be returned

### Requirement: InstallListener is concurrency-safe
The listeners slice SHALL be protected by a mutex, allowing concurrent calls to `InstallListener`.

#### Scenario: Concurrent InstallListener calls
- **WHEN** multiple goroutines call `InstallListener` simultaneously
- **THEN** all successful listeners SHALL be recorded without data race

### Requirement: Service Stop stops listeners first
`baseService.Stop()` SHALL call `stopListeners()` BEFORE cancelling the context and waiting for goroutines. `stopListeners()` iterates all recorded listeners and calls `ln.Stop()` on each.

#### Scenario: Listeners stopped before worker shutdown
- **WHEN** `Stop()` is called on a Service with 2 listeners and 1 component
- **THEN** the execution order SHALL be: stop 2 listeners → cancel context → wait for goroutines → Runner.Shutdown → disconnect 1 component

#### Scenario: Listeners stopped before context cancellation
- **WHEN** `Stop()` is called on a Service
- **THEN** all listeners SHALL be stopped before `cancel()` is called

### Requirement: Service GetMetadata includes listener info
`baseService.GetMetadata()` SHALL populate the Listeners field in WorkerMetaData with the metadata of all installed listeners.

#### Scenario: GetMetadata reflects installed listeners
- **WHEN** 2 listeners have been installed via `InstallListener`
- **THEN** `GetMetadata().Listeners` SHALL contain the ListenerMetaInfo of both listeners

#### Scenario: GetMetadata reflects service labels
- **WHEN** a Service is created with Labels `{"team": "backend"}`
- **THEN** `GetMetadata().Labels` SHALL equal `{"team": "backend"}`
