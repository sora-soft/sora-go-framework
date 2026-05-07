## MODIFIED Requirements

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
