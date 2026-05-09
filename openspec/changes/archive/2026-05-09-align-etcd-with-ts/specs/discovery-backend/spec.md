## MODIFIED Requirements

### Requirement: Backend connection lifecycle
Backend SHALL provide `Connect(ctx)` and `Disconnect()` methods for managing the underlying store connection. Connect SHALL ensure the underlying component is started before use (calling `Start(ctx)` with reference counting). Connect SHALL obtain the leaseID from the EtcdComponent (via `LeaseID()`) rather than creating its own lease. Connect SHALL NOT create, grant, or manage leases directly—lease lifecycle is fully owned by EtcdComponent. Disconnect SHALL cancel watchers and release the component reference by calling `Stop()`, but SHALL NOT revoke or close leases. All entity registrations made while connected SHALL be cleaned up on disconnect.

#### Scenario: Connect establishes store connection
- **WHEN** `Connect(ctx)` is called
- **THEN** the Backend SHALL call `Start(ctx)` on the underlying component, obtain the leaseID from the component's `LeaseID()` method, then establish watchers and perform initial sync using that leaseID

#### Scenario: Connect when component already started
- **WHEN** `Connect(ctx)` is called and the underlying component has already been started by another consumer
- **THEN** the Backend SHALL increment the component's reference count via `Start(ctx)` without re-establishing the physical connection, and proceed normally

#### Scenario: Connect with start failure
- **WHEN** `Connect(ctx)` is called and `Start(ctx)` on the component fails
- **THEN** the Backend SHALL return the error without attempting watcher setup

#### Scenario: Disconnect cleans up
- **WHEN** `Disconnect()` is called
- **THEN** all active watchers SHALL be closed (using cancellable context), the underlying component's `Stop()` SHALL be called to release the reference count, but the lease SHALL NOT be revoked or closed (lease lifecycle is owned by EtcdComponent)

### Requirement: Transparent lease management
Backend implementations SHALL NOT create or manage leases directly. Registered entities SHALL be written using the leaseID provided by the underlying EtcdComponent. Lease lifecycle (grant, keepalive, lost detection, reconnect) is fully owned by EtcdComponent. The Backend SHALL register an `OnLeaseReconnect` callback to re-register local entities when the lease is re-established.

#### Scenario: Entity expires on disconnect
- **WHEN** a service is registered and then the EtcdComponent's lease expires (e.g., etcd cluster unreachable)
- **THEN** the service SHALL no longer be discoverable via the store until reconnect re-registers it

#### Scenario: Entity re-registered on reconnect
- **WHEN** the EtcdComponent reconnects with a new lease
- **THEN** the Backend SHALL re-register all local entities with the new leaseID

### Requirement: Endpoint update validates service association
When the store receives an endpoint update via watcher, it SHALL check whether the endpoint's `TargetID` corresponds to an existing service in the store. If the service does not exist, the endpoint update SHALL be silently skipped.

#### Scenario: Endpoint with existing service
- **WHEN** an endpoint update is received with TargetID="svc-1" and service "svc-1" exists in the store
- **THEN** the endpoint SHALL be stored normally

#### Scenario: Endpoint without existing service
- **WHEN** an endpoint update is received with TargetID="svc-999" and service "svc-999" does not exist in the store
- **THEN** the endpoint SHALL be silently discarded (not stored)

### Requirement: Watcher uses cancellable context
All watchers SHALL be created with a context that can be cancelled during `Disconnect()`. When `Disconnect()` is called, the watcher context SHALL be cancelled, causing all watch goroutines to terminate.

#### Scenario: Watchers stop on disconnect
- **WHEN** `Disconnect()` is called
- **THEN** the watcher context SHALL be cancelled and all watch goroutines SHALL terminate
