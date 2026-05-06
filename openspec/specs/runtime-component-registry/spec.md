### Requirement: Runtime SHALL register components by name
Runtime SHALL provide `RegisterComponent(name string, c component.Component)` that stores the Component in an internal map keyed by name. Registration SHALL NOT call `Start()` on the Component.

#### Scenario: Register stores component without starting
- **WHEN** `rt.RegisterComponent("etcd", comp)` is called
- **THEN** `rt.GetComponent("etcd")` SHALL return `(comp, true)`
- **AND** `comp.Start()` SHALL NOT have been called

#### Scenario: Register with duplicate name overwrites
- **WHEN** `rt.RegisterComponent("etcd", compA)` then `rt.RegisterComponent("etcd", compB)` is called
- **THEN** `rt.GetComponent("etcd")` SHALL return `(compB, true)`

### Requirement: Runtime SHALL query components by name
Runtime SHALL provide `GetComponent(name string) (component.Component, bool)` that returns the registered Component and true, or nil and false if not found.

#### Scenario: Query existing component
- **WHEN** component "etcd" has been registered
- **THEN** `rt.GetComponent("etcd")` SHALL return the registered Component and true

#### Scenario: Query non-existent component
- **WHEN** no component named "redis" has been registered
- **THEN** `rt.GetComponent("redis")` SHALL return nil and false

### Requirement: Component registry SHALL be concurrency-safe
All operations on the component registry SHALL be protected by `sync.RWMutex`.

#### Scenario: Concurrent register and get
- **WHEN** goroutine A calls `RegisterComponent` and goroutine B calls `GetComponent` simultaneously
- **THEN** both operations SHALL complete without data race
