### Requirement: Runtime.Startup accepts NodeRunner and Backend
Runtime SHALL provide `Startup(ctx context.Context, node *NodeRunner, backend discovery.Backend) error` that connects the backend and stores references to the node and backend.

#### Scenario: Successful startup
- **WHEN** `rt.Startup(ctx, nodeRunner, backend)` is called with valid arguments
- **THEN** `backend.Connect(ctx)` SHALL be called
- **AND** `rt.GetNode()` SHALL return the provided nodeRunner
- **AND** `rt.GetBackend()` SHALL return the provided backend

#### Scenario: Backend Connect failure
- **WHEN** `rt.Startup(ctx, nodeRunner, backend)` is called and `backend.Connect(ctx)` returns an error
- **THEN** Startup SHALL return that error
- **AND** `rt.GetNode()` SHALL return nil
- **AND** `rt.GetBackend()` SHALL return nil

### Requirement: Runtime.Startup stores references after successful Connect
Runtime SHALL store the node and backend references only after `backend.Connect(ctx)` succeeds.

#### Scenario: References not stored on Connect failure
- **WHEN** `backend.Connect(ctx)` fails
- **THEN** `rt.GetNode()` SHALL return nil
- **AND** `rt.GetBackend()` SHALL return nil

### Requirement: Runtime SHALL provide GetNode accessor
Runtime SHALL provide `GetNode() *NodeRunner` that returns the stored NodeRunner reference. SHALL be concurrency-safe.

#### Scenario: GetNode returns stored NodeRunner
- **WHEN** `rt.Startup(ctx, nodeRunner, backend)` has succeeded
- **THEN** `rt.GetNode()` SHALL return nodeRunner

#### Scenario: GetNode returns nil before Startup
- **WHEN** Startup has not been called
- **THEN** `rt.GetNode()` SHALL return nil

### Requirement: Runtime SHALL provide GetBackend accessor
Runtime SHALL provide `GetBackend() discovery.Backend` that returns the stored Backend reference. SHALL be concurrency-safe.

#### Scenario: GetBackend returns stored Backend
- **WHEN** `rt.Startup(ctx, nodeRunner, backend)` has succeeded
- **THEN** `rt.GetBackend()` SHALL return backend

#### Scenario: GetBackend returns nil before Startup
- **WHEN** Startup has not been called
- **THEN** `rt.GetBackend()` SHALL return nil

### Requirement: Runtime SHALL provide GetDiscovery accessor
Runtime SHALL provide `GetDiscovery() discovery.Discovery` that returns `backend.Discovery()` if a backend is stored, or nil otherwise. SHALL be concurrency-safe.

#### Scenario: GetDiscovery returns Discovery from backend
- **WHEN** `rt.Startup(ctx, nodeRunner, backend)` has succeeded
- **THEN** `rt.GetDiscovery()` SHALL return `backend.Discovery()`

#### Scenario: GetDiscovery returns nil before Startup
- **WHEN** Startup has not been called
- **THEN** `rt.GetDiscovery()` SHALL return nil

#### Scenario: GetDiscovery returns nil after Shutdown
- **WHEN** Shutdown has cleared the backend reference
- **THEN** `rt.GetDiscovery()` SHALL return nil
