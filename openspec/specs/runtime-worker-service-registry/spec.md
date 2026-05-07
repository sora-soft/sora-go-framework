### Requirement: Runtime SHALL install Service by id
Runtime SHALL provide `InstallService(svc runner.Service)` that stores the Service in an internal map keyed by `svc.GetMetadata().Id`.

#### Scenario: Install stores service
- **WHEN** `rt.InstallService(svc)` is called where `svc.GetMetadata().Id` is "uuid-1"
- **THEN** the service registry SHALL contain key "uuid-1" mapping to svc

### Requirement: Runtime SHALL install Worker by id
Runtime SHALL provide `InstallWorker(w runner.Worker)` that stores the Worker in an internal map keyed by `w.GetMetadata().Id`.

#### Scenario: Install stores worker
- **WHEN** `rt.InstallWorker(w)` is called where `w.GetMetadata().Id` is "uuid-2"
- **THEN** the worker registry SHALL contain key "uuid-2" mapping to w

### Requirement: Runtime SHALL uninstall Service with Stop
Runtime SHALL provide `UninstallService(id string) error` that calls `Stop()` on the Service, removes it from the map, and returns any error from Stop. Removal SHALL occur regardless of whether Stop returns an error.

#### Scenario: Successful uninstall
- **WHEN** `rt.UninstallService("uuid-1")` is called for a running service
- **THEN** `svc.Stop()` SHALL be called, the entry SHALL be removed from the map, and error SHALL be nil

#### Scenario: Uninstall with Stop failure
- **WHEN** `rt.UninstallService("uuid-1")` is called and `svc.Stop()` returns an error
- **THEN** the entry SHALL still be removed from the map, and the error SHALL be returned

#### Scenario: Uninstall non-existent service
- **WHEN** `rt.UninstallService("non-existent")` is called
- **THEN** it SHALL return without error (no-op)

### Requirement: Runtime SHALL uninstall Worker with Stop
Runtime SHALL provide `UninstallWorker(id string) error` that calls `Stop()` on the Worker, removes it from the map, and returns any error from Stop. Removal SHALL occur regardless of whether Stop returns an error.

#### Scenario: Successful uninstall
- **WHEN** `rt.UninstallWorker("uuid-2")` is called for a running worker
- **THEN** `w.Stop()` SHALL be called, the entry SHALL be removed from the map, and error SHALL be nil

#### Scenario: Uninstall non-existent worker
- **WHEN** `rt.UninstallWorker("non-existent")` is called
- **THEN** it SHALL return without error (no-op)

### Requirement: Worker/Service registry SHALL be concurrency-safe
All operations on both registries SHALL be protected by `sync.RWMutex`.

#### Scenario: Concurrent install and uninstall
- **WHEN** goroutine A calls `InstallService` and goroutine B calls `UninstallService` simultaneously
- **THEN** both operations SHALL complete without data race

### Requirement: Runtime SHALL provide Startup with node and backend
Runtime SHALL provide `Startup(ctx context.Context, node *NodeRunner, backend discovery.Backend) error` that connects the backend and stores references. This replaces the previous no-op `Startup() error`.

#### Scenario: Startup connects backend and stores references
- **WHEN** `rt.Startup(ctx, nodeRunner, backend)` is called
- **THEN** `backend.Connect(ctx)` SHALL be called
- **AND** references SHALL be stored and accessible via GetNode/GetBackend/GetDiscovery

### Requirement: Runtime SHALL provide Shutdown with graceful teardown
Runtime SHALL provide `Shutdown() error` that concurrently stops all Services and Workers (excluding the node service), then stops the node, then disconnects the backend. This replaces the previous no-op `Shutdown() error`.

#### Scenario: Shutdown performs full graceful teardown
- **WHEN** `rt.Shutdown()` is called after Startup
- **THEN** all Services and Workers SHALL be stopped concurrently
- **AND** the node SHALL be stopped after all others complete
- **AND** backend.Disconnect() SHALL be called last
