## 1. Package Structure & Metadata Types

- [x] 1.1 Create `pkg/discovery/` package directory structure
- [x] 1.2 Define `NodeMeta`, `ServiceMeta`, `WorkerMeta`, `EndpointMeta` structs in `pkg/discovery/metadata.go` with JSON tags matching TS interfaces
- [x] 1.3 Define `BackendInfo` struct in `pkg/discovery/metadata.go`
- [x] 1.4 Implement `NewEndpointMetaFromListener()` conversion function in `pkg/discovery/metadata.go`

## 2. Core Interfaces

- [x] 2.1 Define `Registry` interface in `pkg/discovery/registry.go` (Register/Unregister for Node, Service, Endpoint, Worker)
- [x] 2.2 Define `Discovery` interface in `pkg/discovery/discovery.go` (Get/List/Watch for all entity types, including ListServicesByName, ListWorkersByName, ListEndpointsByService)
- [x] 2.3 Define `Election` interface in `pkg/discovery/election.go` (Campaign, Resign, Leader, Watch)
- [x] 2.4 Define `Backend` interface in `pkg/discovery/backend.go` (Connect, Disconnect, Registry, Discovery, NewElection, GetInfo)

## 3. RAM Store Implementation

- [x] 3.1 Create `pkg/discovery/store/ram/` package with internal store struct using `sync.RWMutex` and `map[string]T` for each entity type
- [x] 3.2 Implement RAM `Registry` — register/unregister for all entity types, trigger snapshot push to watchers on each mutation
- [x] 3.3 Implement RAM `Discovery` — get by ID, list all, list by name/serviceID using map lookups and filtering
- [x] 3.4 Implement RAM `Watch` — return channel, push initial snapshot, subscribe to mutation notifications, close on context cancel
- [x] 3.5 Implement RAM `Election` — campaign blocks if leader exists, first candidate wins, resign releases leadership and unblocks next candidate
- [x] 3.6 Implement RAM `Backend` — Connect/Disconnect no-ops, return Registry/Discovery singletons, NewElection creates instances sharing a leader map by name
- [x] 3.7 Write tests for RAM Registry round-trip (register → get → unregister → get nil)
- [x] 3.8 Write tests for RAM Watch (initial snapshot, update on register, update on unregister, context cancel closes channel)
- [x] 3.9 Write tests for RAM Election (first campaign wins, second blocks, resign transfers leadership)

## 4. Runner Integration

- [x] 4.1 Add optional `Registry` field to `baseService` in `pkg/runner/service.go`
- [x] 4.2 In `baseService.Start()` — after listeners are started and state is Ready, call `RegisterService` then `RegisterEndpoint` for each listener (with TargetID/TargetName populated) if Registry is set
- [x] 4.3 In `baseService.Stop()` — before shutting down listeners, call `UnregisterEndpoint` for each endpoint then `UnregisterService` if Registry is set
- [x] 4.4 Store registered endpoint IDs on the service for correct unregistration during stop
- [x] 4.5 Write integration test: start service with RAM backend → verify entity appears in Discovery → stop service → verify entity removed
