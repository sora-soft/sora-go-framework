## ADDED Requirements

### Requirement: Backend connection lifecycle
Backend SHALL provide `Connect(ctx)` and `Disconnect()` methods for managing the underlying store connection. All entity registrations made while connected SHALL be cleaned up on disconnect.

#### Scenario: Connect establishes store connection
- **WHEN** `Connect(ctx)` is called
- **THEN** the Backend SHALL establish a connection to the underlying store and make Registry/Discovery operations functional

#### Scenario: Disconnect cleans up
- **WHEN** `Disconnect()` is called
- **THEN** all active watchers SHALL be closed and the underlying store connection SHALL be released

### Requirement: Backend provides Registry
Backend SHALL provide a `Registry()` method that returns a `Registry` interface instance.

#### Scenario: Registry returns functional Registry
- **WHEN** `Registry()` is called after `Connect()`
- **THEN** it SHALL return a Registry instance capable of register/unregister operations

### Requirement: Backend provides Discovery
Backend SHALL provide a `Discovery()` method that returns a `Discovery` interface instance.

#### Scenario: Discovery returns functional Discovery
- **WHEN** `Discovery()` is called after `Connect()`
- **THEN** it SHALL return a Discovery instance capable of get/list/watch operations

### Requirement: Backend creates Election instances
Backend SHALL provide a `NewElection(name)` method that returns an `Election` interface instance. Each call with the same name SHALL return an election operating on the same logical leader slot.

#### Scenario: NewElection returns Election instance
- **WHEN** `NewElection("singleton-job")` is called
- **THEN** it SHALL return an Election instance for the "singleton-job" leader slot

#### Scenario: Same name shares leader slot
- **WHEN** `NewElection("singleton-job")` is called twice and the first instance campaigns and wins
- **THEN** the second instance's `Leader()` SHALL return the first instance's campaign ID

### Requirement: Backend exposes type info
Backend SHALL provide a `GetInfo()` method returning a `BackendInfo` struct with `Type` (e.g., "ram", "etcd", "zookeeper") and `Version` fields.

#### Scenario: GetInfo returns backend identity
- **WHEN** `GetInfo()` is called on a ram backend
- **THEN** it SHALL return `BackendInfo{Type: "ram", Version: "0.0.0"}`

### Requirement: Transparent lease management
Backend implementations SHALL handle lease creation and heartbeat internally. Registered entities SHALL be automatically removed when the Backend connection is lost or the underlying lease expires. The caller SHALL NOT be exposed to lease lifecycle details.

#### Scenario: Entity expires on disconnect
- **WHEN** a service is registered and then `Disconnect()` is called without first unregistering the service
- **THEN** the service SHALL no longer be discoverable via the store
