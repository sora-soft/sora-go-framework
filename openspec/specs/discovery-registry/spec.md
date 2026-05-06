### Requirement: Register and unregister node
Registry SHALL provide `RegisterNode(ctx, NodeMeta)` and `UnregisterNode(ctx, id)` methods. Registering a node SHALL store its metadata in the discovery store. Unregistering SHALL remove it.

#### Scenario: Register a node
- **WHEN** `RegisterNode(ctx, nodeMeta)` is called with valid node metadata
- **THEN** the node SHALL be stored and retrievable via `Discovery.GetNode(ctx, id)`

#### Scenario: Unregister a node
- **WHEN** `UnregisterNode(ctx, id)` is called with the ID of a previously registered node
- **THEN** the node SHALL be removed and `Discovery.GetNode(ctx, id)` SHALL return nil

### Requirement: Register and unregister service
Registry SHALL provide `RegisterService(ctx, ServiceMeta)` and `UnregisterService(ctx, id)` methods.

#### Scenario: Register a service
- **WHEN** `RegisterService(ctx, serviceMeta)` is called with valid service metadata
- **THEN** the service SHALL be stored and retrievable via `Discovery.GetService(ctx, id)`

#### Scenario: Unregister a service
- **WHEN** `UnregisterService(ctx, id)` is called with the ID of a previously registered service
- **THEN** the service SHALL be removed and `Discovery.GetService(ctx, id)` SHALL return nil

### Requirement: Register and unregister endpoint
Registry SHALL provide `RegisterEndpoint(ctx, EndpointMeta)` and `UnregisterEndpoint(ctx, id)` methods. The `EndpointMeta.TargetID` and `EndpointMeta.TargetName` fields SHALL be populated by the caller before registration.

#### Scenario: Register an endpoint
- **WHEN** `RegisterEndpoint(ctx, endpointMeta)` is called with valid endpoint metadata where `TargetID` and `TargetName` are set
- **THEN** the endpoint SHALL be stored and retrievable via `Discovery.GetEndpoint(ctx, id)`

#### Scenario: Unregister an endpoint
- **WHEN** `UnregisterEndpoint(ctx, id)` is called with the ID of a previously registered endpoint
- **THEN** the endpoint SHALL be removed and `Discovery.GetEndpoint(ctx, id)` SHALL return nil

### Requirement: Register and unregister worker
Registry SHALL provide `RegisterWorker(ctx, WorkerMeta)` and `UnregisterWorker(ctx, id)` methods.

#### Scenario: Register a worker
- **WHEN** `RegisterWorker(ctx, workerMeta)` is called with valid worker metadata
- **THEN** the worker SHALL be stored and retrievable via `Discovery.GetWorker(ctx, id)`

#### Scenario: Unregister a worker
- **WHEN** `UnregisterWorker(ctx, id)` is called with the ID of a previously registered worker
- **THEN** the worker SHALL be removed and `Discovery.GetWorker(ctx, id)` SHALL return nil
