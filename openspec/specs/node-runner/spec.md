### Requirement: NodeRunner implements Runner interface
NodeRunner SHALL implement the `Runner` interface with `Startup(ctx context.Context) error` and `Shutdown() error` methods.

#### Scenario: NodeRunner satisfies Runner interface
- **WHEN** a `NodeRunner` instance is created
- **THEN** it SHALL be assignable to a `Runner` interface variable

### Requirement: NodeRunner implements ServiceRefAware
NodeRunner SHALL implement the `ServiceRefAware` interface with `SetServiceRef(svcRef ServiceRef)`. The injected ServiceRef SHALL be used during Startup to install Listeners.

#### Scenario: NewService injects ServiceRef into NodeRunner
- **WHEN** `NewService("node", nodeRunner, opts)` is called
- **THEN** the `ServiceRefAware.SetServiceRef` SHALL be called with the ServiceRef

### Requirement: NodeRunner installs Listeners during Startup
NodeRunner SHALL accept a slice of `*rpc.Listener` at construction. During `Startup`, NodeRunner SHALL iterate over the listeners and call `svcRef.InstallListener(ctx, l)` for each one.

#### Scenario: Startup installs all pre-configured listeners
- **WHEN** `Startup(ctx)` is called on NodeRunner with 3 listeners configured
- **THEN** `svcRef.InstallListener` SHALL be called exactly 3 times, once per listener

#### Scenario: Startup installs zero listeners gracefully
- **WHEN** `Startup(ctx)` is called on NodeRunner with no listeners configured
- **THEN** the Startup SHALL complete successfully without error

#### Scenario: Listener install failure aborts Startup
- **WHEN** `svcRef.InstallListener` returns an error for any listener
- **THEN** Startup SHALL return that error immediately without installing remaining listeners

### Requirement: NodeRunner registers NodeId during Startup
NodeRunner SHALL call `runtime.RT.SetNodeId(svc.GetMetadata().Id)` during Startup after installing listeners. The NodeId SHALL be the ID of the wrapping Service.

#### Scenario: Startup sets NodeId on Runtime
- **WHEN** `Startup(ctx)` completes successfully
- **THEN** `runtime.RT.NodeId()` SHALL return the Service's ID

### Requirement: NodeRunner Shutdown is no-op
NodeRunner's `Shutdown()` SHALL return nil without performing any cleanup. Listener and component lifecycle are managed by the wrapping Service.

#### Scenario: Shutdown returns nil
- **WHEN** `Shutdown()` is called on NodeRunner
- **THEN** it SHALL return nil

### Requirement: NodeRunner StateData method
NodeRunner SHALL provide a `StateData() NodeMetaData` method that returns the current node state metadata, including Id, Alias, Host, Pid, Endpoints, State, StartTime, and Versions.

#### Scenario: StateData returns populated NodeMetaData
- **WHEN** `StateData()` is called after successful Startup
- **THEN** the returned NodeMetaData SHALL contain the Service ID, configured Alias, current hostname, current PID, listener endpoints, current state, start timestamp, and version strings

### Requirement: NodeRunner RunData method
NodeRunner SHALL provide a `RunData() NodeRunData` method that aggregates the NodeMetaData with metadata from Runtime's registered Components, Services, and Workers.

#### Scenario: RunData returns complete process snapshot
- **WHEN** `RunData()` is called after components and workers are registered in Runtime
- **THEN** the returned NodeRunData SHALL contain the Node's StateData plus ComponentMetadata for all registered components and WorkerMetaData for all installed workers and services

### Requirement: NodeRunner holds Service reference
NodeRunner SHALL hold a reference to the wrapping `Service` to read its ID and metadata. This reference SHALL be set during Startup or via a separate setter method.

#### Scenario: NodeRunner reads Service ID for NodeId registration
- **WHEN** NodeRunner needs to register the NodeId
- **THEN** it SHALL read the ID from the wrapping Service's GetMetadata().Id
