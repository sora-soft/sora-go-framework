## MODIFIED Requirements

### Requirement: Backend connection lifecycle
Backend SHALL provide `Connect(ctx)` and `Disconnect()` methods for managing the underlying store connection. Connect SHALL ensure the underlying component is started before use (calling `Start(ctx)` with reference counting). Disconnect SHALL release the component reference by calling `Stop()` after cleaning up watchers and lease. All entity registrations made while connected SHALL be cleaned up on disconnect.

#### Scenario: Connect establishes store connection
- **WHEN** `Connect(ctx)` is called
- **THEN** the Backend SHALL call `Start(ctx)` on the underlying component to ensure it is connected, then establish the store session (lease, watchers, initial sync) and make Registry/Discovery operations functional

#### Scenario: Connect when component already started
- **WHEN** `Connect(ctx)` is called and the underlying component has already been started by another consumer
- **THEN** the Backend SHALL increment the component's reference count via `Start(ctx)` without re-establishing the physical connection, and proceed normally

#### Scenario: Connect with start failure
- **WHEN** `Connect(ctx)` is called and `Start(ctx)` on the component fails
- **THEN** the Backend SHALL return the error without attempting lease or watcher setup

#### Scenario: Disconnect cleans up
- **WHEN** `Disconnect()` is called
- **THEN** all active watchers SHALL be closed, the lease SHALL be revoked, and then `Stop()` SHALL be called on the underlying component to release the reference count
