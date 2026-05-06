## Why

The framework currently has no mechanism for cross-process service registration and discovery. Services, workers, and endpoints exist only within a single process's runtime. To support distributed deployment (multiple nodes forming a cluster), we need a service discovery layer that allows any node to register its entities and discover entities on other nodes. This is critical for RPC connector routing — a client on Node A needs to discover which endpoints a service exposes on Node B. The cluster is a mixed TypeScript/Go environment, so the wire format must be compatible with the existing TS `Discovery` implementation.

## What Changes

- Add a `pkg/discovery` package with independent data models (`NodeMeta`, `ServiceMeta`, `WorkerMeta`, `EndpointMeta`) aligned with the TS `IListenerMetaData`/`IServiceMetaData`/`IWorkerMetaData`/`INodeMetaData` interfaces
- Add `Registry` interface for write-side operations (register/unregister all four entity types)
- Add `Discovery` interface for read-side operations (get/list by ID or name, watch with full-snapshot channel push)
- Add `Election` interface for leader election (campaign, resign, leader query, watch)
- Add `Backend` interface as the entry point that creates Registry, Discovery, and Election instances sharing the same underlying connection
- Add a `store/ram` in-memory implementation for testing (mirrors TS `RamDiscovery`/`RamElection`)
- Integrate service/endpoint registration into the existing `ServiceRunner` lifecycle
- Extend `EndpointMeta` with `ID`, `Weight`, `TargetID`, `TargetName` fields (TS compatibility); `TargetID`/`TargetName` auto-populated by the service layer at registration time

## Capabilities

### New Capabilities
- `discovery-registry`: Write-side interface for registering and unregistering nodes, services, workers, and endpoints in the discovery store
- `discovery-query`: Read-side interface for querying and watching entities (get by ID, list all, list by name, watch with full-snapshot push)
- `discovery-election`: Leader election interface with campaign, resign, leader query, and leader change watch
- `discovery-backend`: Pluggable backend entry point that manages the underlying store connection and provides Registry, Discovery, and Election instances with transparent lease/heartbeat handling
- `discovery-metadata`: Independent data model types for service discovery, JSON-serializable and wire-compatible with the TypeScript `Discovery` implementation

### Modified Capabilities
- `worker-lifecycle`: Service startup/shutdown will include registration/unregistration of service and its endpoints with the discovery Registry

## Impact

- **New package**: `pkg/discovery/` with interfaces, metadata types, and store implementations
- **Modified**: `pkg/runner/service.go` — ServiceRunner startup registers service + endpoints; shutdown unregisters them
- **Dependencies**: No new external dependencies for interfaces/metadata; `store/ram` uses only stdlib; future `store/etcd`/`store/zookeeper` will add respective clients
- **Wire format**: JSON field names must exactly match TS interface property names for cross-language cluster compatibility
