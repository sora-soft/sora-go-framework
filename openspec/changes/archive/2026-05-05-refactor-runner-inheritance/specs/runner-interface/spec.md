## ADDED Requirements

### Requirement: Worker public interface
`Worker` SHALL be the public interface for worker consumers, exposing `Start(ctx context.Context) error`, `Stop() error`, `Go(fn func(ctx context.Context))`, and `GetMetadata() WorkerMetaData`.

#### Scenario: NewWorker returns Worker interface
- **WHEN** `NewWorker(name, runner, opts)` is called
- **THEN** it SHALL return a `Worker` interface value

#### Scenario: Worker interface methods are callable
- **WHEN** a consumer holds a `Worker` interface value
- **THEN** it SHALL be able to call `Start`, `Stop`, `Go`, and `GetMetadata`

### Requirement: Service public interface extends Worker
`Service` SHALL be a public interface that embeds `Worker`, establishing Service is a Worker. No additional methods are required beyond Worker's.

#### Scenario: NewService returns Service interface
- **WHEN** `NewService(name, runner, opts)` is called
- **THEN** it SHALL return a `Service` interface value

#### Scenario: Service is assignable to Worker
- **WHEN** a `Service` interface value is assigned to a `Worker` variable
- **THEN** the assignment SHALL compile without error

#### Scenario: Service GetMetadata includes extended fields
- **WHEN** `GetMetadata()` is called on a `Service` value
- **THEN** the returned `WorkerMetaData` SHALL include populated Labels and Listeners fields

### Requirement: Runner hook interface unchanged
`Runner` SHALL remain the hook interface with `Startup(ctx context.Context) error` and `Shutdown() error`. External implementors only need to implement Runner.

#### Scenario: Runner interface has exactly two methods
- **WHEN** an external type implements `Startup` and `Shutdown`
- **THEN** it SHALL satisfy the `Runner` interface

### Requirement: baseWorker is unexported
`baseWorker` struct SHALL NOT be exported. It SHALL implement both `Worker` and `WorkerRef` interfaces.

#### Scenario: External package cannot reference baseWorker
- **WHEN** external code attempts to use `runner.baseWorker`
- **THEN** the compiler SHALL reject the reference

### Requirement: baseService is unexported
`baseService` struct SHALL NOT be exported. It SHALL implement both `Service` and `ServiceRef` interfaces.

#### Scenario: External package cannot reference baseService
- **WHEN** external code attempts to use `runner.baseService`
- **THEN** the compiler SHALL reject the reference

### Requirement: WorkerMetaData unified structure
`WorkerMetaData` SHALL contain all metadata fields for both Worker and Service, including `Labels utility.Labels` and `Listeners []rpc.ListenerMetaInfo`. Worker instances SHALL leave these fields empty (zero value).

#### Scenario: Worker GetMetadata returns empty Labels and Listeners
- **WHEN** `GetMetadata()` is called on a Worker (not Service)
- **THEN** Labels SHALL be nil/empty and Listeners SHALL be nil/empty

#### Scenario: Service GetMetadata returns populated Labels and Listeners
- **WHEN** `GetMetadata()` is called on a Service
- **THEN** Labels SHALL contain the service's labels and Listeners SHALL contain metadata of all installed listeners

#### Scenario: WorkerMetaData JSON omits empty fields
- **WHEN** WorkerMetaData is serialized to JSON with empty Labels and Listeners
- **THEN** those fields SHALL NOT appear in the JSON output
