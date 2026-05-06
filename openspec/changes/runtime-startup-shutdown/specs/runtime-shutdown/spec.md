## ADDED Requirements

### Requirement: Runtime.Shutdown stops all Services and Workers concurrently
Runtime SHALL stop all registered Services and Workers concurrently using goroutines and a WaitGroup, excluding the node service. Shutdown SHALL collect all IDs first, then launch concurrent stop operations.

#### Scenario: Shutdown stops all Services and Workers
- **WHEN** `rt.Shutdown()` is called with 2 Services and 3 Workers registered
- **THEN** all 5 `Stop()` calls SHALL be invoked concurrently
- **AND** Shutdown SHALL wait for all to complete before proceeding

#### Scenario: Shutdown with no Services or Workers
- **WHEN** `rt.Shutdown()` is called with empty service and worker registries
- **THEN** Shutdown SHALL proceed directly to node shutdown without error

### Requirement: Runtime.Shutdown collects errors from all stop operations
Runtime SHALL collect errors from all Service/Worker stop operations. If any error occurs, Shutdown SHALL return the first error encountered.

#### Scenario: One Service Stop fails
- **WHEN** `rt.Shutdown()` is called and one Service.Stop() returns an error
- **THEN** Shutdown SHALL still stop all remaining Services and Workers
- **AND** Shutdown SHALL return the error from the failed Service

#### Scenario: Multiple stops fail
- **WHEN** `rt.Shutdown()` is called and multiple stop operations return errors
- **THEN** Shutdown SHALL return the first error
- **AND** all stop operations SHALL still be attempted

### Requirement: Runtime.Shutdown stops node after all Services and Workers
Runtime SHALL stop the node service after all Services and Workers have completed their stop operations. The node SHALL be stopped via `UninstallService` using the node's service ID stored in `rt.NodeId()`.

#### Scenario: Node stopped after Workers and Services
- **WHEN** `rt.Shutdown()` is called with Services, Workers, and a node registered
- **THEN** all non-node Services and Workers SHALL be stopped first
- **AND** the node service SHALL be stopped after all others have completed

#### Scenario: Node shutdown when nodeId is empty
- **WHEN** `rt.Shutdown()` is called and `rt.NodeId()` returns empty string
- **THEN** node stop SHALL be skipped

### Requirement: Runtime.Shutdown disconnects Backend last
Runtime SHALL disconnect the discovery Backend after all Services, Workers, and the node have been stopped. Backend.Disconnect() SHALL be called regardless of whether previous stop operations returned errors.

#### Scenario: Backend disconnected after all stops
- **WHEN** `rt.Shutdown()` completes the Service/Worker/Node stop phase
- **THEN** `backend.Disconnect()` SHALL be called

#### Scenario: Backend Disconnect error combined with stop errors
- **WHEN** a Service.Stop() fails and backend.Disconnect() also fails
- **THEN** Shutdown SHALL return the first error encountered (Service error)

#### Scenario: Backend Disconnect failure does not prevent reference cleanup
- **WHEN** `backend.Disconnect()` returns an error
- **THEN** `rt.GetBackend()` SHALL still return nil (references cleared)

### Requirement: Runtime.Shutdown clears stored references
Runtime SHALL clear the stored node and backend references after calling Disconnect.

#### Scenario: References cleared after Shutdown
- **WHEN** `rt.Shutdown()` completes
- **THEN** `rt.GetNode()` SHALL return nil
- **AND** `rt.GetBackend()` SHALL return nil
- **AND** `rt.GetDiscovery()` SHALL return nil

### Requirement: Runtime.Shutdown is safe when not started
Runtime SHALL handle Shutdown being called without a prior Startup without panicking.

#### Scenario: Shutdown without Startup
- **WHEN** `rt.Shutdown()` is called without ever calling Startup
- **THEN** Shutdown SHALL return nil without panicking

### Requirement: Runtime.Shutdown skip service matching nodeId
When collecting service IDs for concurrent stop, Runtime SHALL exclude the service whose ID matches `rt.NodeId()` from the concurrent phase. That service SHALL be stopped in the node stop phase.

#### Scenario: Node service not stopped concurrently
- **WHEN** `rt.Shutdown()` is called and nodeId is "node-uuid"
- **AND** services map contains a service with ID "node-uuid"
- **THEN** the concurrent stop phase SHALL NOT include the service "node-uuid"
- **AND** the node stop phase SHALL stop service "node-uuid"
