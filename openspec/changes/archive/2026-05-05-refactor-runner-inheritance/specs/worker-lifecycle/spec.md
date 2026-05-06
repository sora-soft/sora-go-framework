## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Service Stop overrides Worker Stop
`baseService.Stop()` SHALL stop all installed listeners before executing the Worker stop sequence (cancel context, wait for goroutines, Runner.Shutdown, disconnect components).

#### Scenario: Service Stop stops listeners then delegates to worker cleanup
- **WHEN** `Stop()` is called on a Service
- **THEN** the sequence SHALL be: set state Stopping → set running false → stopListeners → cancel context → wait for goroutines → Runner.Shutdown → disconnectComponents → set state Stopped
