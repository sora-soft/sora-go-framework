### Requirement: WorkerRef interface
`WorkerRef` SHALL be a public interface providing `ConnectComponent(ctx context.Context, c component.Component) error`. It represents the framework-level capability for component registration.

#### Scenario: WorkerRef has ConnectComponent method
- **WHEN** a type implements `ConnectComponent(ctx context.Context, c component.Component) error`
- **THEN** it SHALL satisfy the `WorkerRef` interface

### Requirement: ServiceRef interface extends WorkerRef
`ServiceRef` SHALL be a public interface that embeds `WorkerRef` and adds `InstallListener(ctx context.Context, l *rpc.Listener) error`.

#### Scenario: ServiceRef has both ConnectComponent and InstallListener
- **WHEN** a type implements `ConnectComponent` and `InstallListener`
- **THEN** it SHALL satisfy the `ServiceRef` interface

### Requirement: WorkerRefAware optional interface
`WorkerRefAware` SHALL be a public interface with `SetWorkerRef(WorkerRef)`. Runner implementations MAY implement this to receive a WorkerRef during construction.

#### Scenario: NewWorker injects WorkerRef when Runner implements WorkerRefAware
- **WHEN** `NewWorker` is called with a Runner that implements `WorkerRefAware`
- **THEN** `SetWorkerRef` SHALL be called with the baseWorker (which implements WorkerRef) before returning

#### Scenario: NewWorker proceeds normally when Runner does not implement WorkerRefAware
- **WHEN** `NewWorker` is called with a Runner that does NOT implement `WorkerRefAware`
- **THEN** the worker SHALL be created and returned without calling SetWorkerRef

### Requirement: ServiceRefAware optional interface
`ServiceRefAware` SHALL be a public interface with `SetServiceRef(ServiceRef)`. Runner implementations MAY implement this to receive a ServiceRef during construction.

#### Scenario: NewService injects ServiceRef when Runner implements ServiceRefAware
- **WHEN** `NewService` is called with a Runner that implements `ServiceRefAware`
- **THEN** `SetServiceRef` SHALL be called with the baseService (which implements ServiceRef) before returning

#### Scenario: NewService falls back to WorkerRef when Runner only implements WorkerRefAware
- **WHEN** `NewService` is called with a Runner that implements `WorkerRefAware` but NOT `ServiceRefAware`
- **THEN** `SetWorkerRef` SHALL be called with the baseWorker embedded in baseService

#### Scenario: NewService proceeds normally when Runner implements neither
- **WHEN** `NewService` is called with a Runner that implements neither `ServiceRefAware` nor `WorkerRefAware`
- **THEN** the service SHALL be created and returned without injecting any ref
