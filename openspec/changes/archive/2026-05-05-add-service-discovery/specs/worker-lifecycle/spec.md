## MODIFIED Requirements

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

## ADDED Requirements

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
