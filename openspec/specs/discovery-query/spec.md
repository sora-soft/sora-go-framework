### Requirement: Get entity by ID
Discovery SHALL provide `GetNode`, `GetService`, `GetWorker`, and `GetEndpoint` methods, each accepting a context and an ID string, returning a pointer to the metadata or nil if not found.

#### Scenario: Get existing service by ID
- **WHEN** `GetService(ctx, id)` is called with the ID of a registered service
- **THEN** it SHALL return a pointer to the `ServiceMeta` with matching ID

#### Scenario: Get non-existent entity by ID
- **WHEN** `GetService(ctx, id)` is called with an ID that does not exist
- **THEN** it SHALL return nil without error

### Requirement: List all entities
Discovery SHALL provide `ListNodes`, `ListServices`, `ListWorkers`, and `ListEndpoints` methods, each returning all entities of that type.

#### Scenario: List all services
- **WHEN** `ListServices(ctx)` is called
- **THEN** it SHALL return a slice containing all registered `ServiceMeta` entries

#### Scenario: List with no entities
- **WHEN** `ListServices(ctx)` is called and no services are registered
- **THEN** it SHALL return an empty (non-nil) slice

### Requirement: List entities by name
Discovery SHALL provide `ListServicesByName(ctx, name)` and `ListWorkersByName(ctx, name)` methods that filter by the entity's `Name` field.

#### Scenario: List services by name with matches
- **WHEN** `ListServicesByName(ctx, "auth")` is called and services named "auth" exist
- **THEN** it SHALL return only services whose `Name` field equals "auth"

#### Scenario: List services by name with no matches
- **WHEN** `ListServicesByName(ctx, "nonexistent")` is called
- **THEN** it SHALL return an empty slice

### Requirement: List endpoints by service ID
Discovery SHALL provide `ListEndpointsByService(ctx, serviceID)` that returns all endpoints whose `TargetID` matches the given service ID.

#### Scenario: List endpoints for a service
- **WHEN** `ListEndpointsByService(ctx, serviceID)` is called with a service ID that has registered endpoints
- **THEN** it SHALL return all `EndpointMeta` entries where `TargetID` equals the given service ID

#### Scenario: List endpoints for a service with none
- **WHEN** `ListEndpointsByService(ctx, serviceID)` is called with a service ID that has no endpoints
- **THEN** it SHALL return an empty slice

### Requirement: Watch entities with full snapshot
Discovery SHALL provide `WatchNodes`, `WatchServices`, `WatchWorkers`, and `WatchEndpoints` methods. Each SHALL return a receive-only channel `<-chan []T`. The channel SHALL receive the current full snapshot immediately upon subscription. On every change to the entity type, the channel SHALL receive the new full snapshot. The channel SHALL close when the context is cancelled or the underlying store encounters a fatal error.

#### Scenario: Watch receives initial snapshot
- **WHEN** `WatchServices(ctx)` is called and services A and B are already registered
- **THEN** the returned channel SHALL immediately yield a slice containing both service A and B

#### Scenario: Watch receives update on registration
- **WHEN** a watcher is active via `WatchServices` and a new service C is registered
- **THEN** the channel SHALL yield a full snapshot containing services A, B, and C

#### Scenario: Watch receives update on unregistration
- **WHEN** a watcher is active and service B is unregistered
- **THEN** the channel SHALL yield a full snapshot containing only the remaining services

#### Scenario: Watch channel closes on context cancellation
- **WHEN** the context passed to `WatchServices(ctx)` is cancelled
- **THEN** the returned channel SHALL be closed
