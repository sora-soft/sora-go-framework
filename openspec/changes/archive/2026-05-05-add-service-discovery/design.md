## Context

The sora-go-framework currently manages services and workers within a single process. The `runner` package provides `NodeRunner`, `ServiceRunner`, and `WorkerRunner` with lifecycle management, and the `rpc` package provides `Listener`/`Connector` for network communication. However, there is no mechanism for a node to advertise its services to the cluster or for a node to discover services on other nodes.

A parallel TypeScript framework (`sora-node`) already implements a `Discovery` abstract class with four entity types (Node, Service, Worker, Endpoint) and an `Election` class for leader election, backed by etcd in production and an in-memory `RamDiscovery` for testing. The Go framework must interoperate with TS nodes in the same cluster — sharing the same discovery store with compatible wire formats.

The existing codebase already defines metadata types that map closely to the TS interfaces: `runner.NodeMetaData`, `types.WorkerMetaData` (used for both services and workers), and `rpc.ListenerMetaInfo`. The discovery package will define its own independent metadata types to avoid circular dependencies, with conversion handled at the integration boundary.

## Goals / Non-Goals

**Goals:**
- Define abstract interfaces for Registry (write-side), Discovery (read-side + watch), Election, and Backend (entry point)
- Define independent data model types in `pkg/discovery` that are JSON-serializable and wire-compatible with TS `Discovery` metadata interfaces
- Provide an in-memory `store/ram` implementation for testing
- Integrate service + endpoint registration into the existing ServiceRunner lifecycle
- Support full-snapshot watch semantics (every change pushes the complete list)
- Support transparent lease/heartbeat handling at the Backend layer

**Non-Goals:**
- etcd or ZooKeeper implementations (will be separate changes)
- Node registration via Runtime startup (deferred to a future change)
- Incremental/delta watch semantics
- Service mesh, load balancing, or circuit breaking (out of scope)
- Health checking beyond lease-based liveness

## Decisions

### 1. Discovery package defines its own metadata types

**Decision**: `pkg/discovery` defines `NodeMeta`, `ServiceMeta`, `WorkerMeta`, `EndpointMeta` as independent structs rather than reusing `runner.NodeMetaData` / `types.WorkerMetaData` / `rpc.ListenerMetaInfo`.

**Rationale**: Avoids circular dependencies (runner → discovery → runner). Keeps the discovery package self-contained and aligned with the wire format. The runner layer performs conversion when calling Registry methods. This also allows the discovery metadata to evolve independently from the internal runtime representations.

**Alternative considered**: Reusing existing types directly — rejected due to tight coupling and the risk of internal fields leaking into the wire format.

### 2. Registry and Discovery as separate interfaces

**Decision**: Split write operations into `Registry` and read/watch operations into `Discovery`.

**Rationale**: Consumers (e.g., RPC connectors looking up endpoints) only need the `Discovery` interface and should not be exposed to registration methods. Producers (service runners) only need `Registry`. This follows the principle of least privilege and makes dependency injection clearer.

**Alternative considered**: Single unified interface (as in TS version) — rejected because Go interfaces work best when small and focused.

### 3. Backend as the composition root

**Decision**: `Backend` is the entry point that manages the underlying store connection and provides `Registry()`, `Discovery()`, and `NewElection(name)`.

**Rationale**: Ensures Registry, Discovery, and Election share the same underlying connection (etcd client, zk client, etc.). The Backend transparently handles lease creation and heartbeat for registered entities. This matches the requirement that Election must use the same backend as Discovery.

### 4. Full-snapshot watch via channels

**Decision**: Each `Watch*` method returns `<-chan []T`. The first message is the current full snapshot; subsequent messages are full snapshots on every change. Channel closes on error or context cancellation.

**Rationale**: Matches the TS `BehaviorSubject` semantics (new subscriber gets current state, then updates). Simple to implement and consume. For typical cluster sizes (tens to hundreds of services), full-snapshot overhead is negligible.

**Alternative considered**: Incremental events (Add/Update/Delete) — rejected for initial implementation due to complexity. Can be layered on later if needed.

### 5. EndpointMeta extends ListenerMetaInfo concept independently

**Decision**: `EndpointMeta` in the discovery package includes all fields from `rpc.ListenerMetaInfo` plus `ID`, `Weight`, `TargetID`, `TargetName`. These fields are not added to `rpc.ListenerMetaInfo`.

**Rationale**: `targetId`/`targetName` are discovery-layer concepts (linking an endpoint to its owning service). The RPC layer should not depend on discovery concepts. The conversion from `ListenerMetaInfo` to `EndpointMeta` happens at the runner boundary when the service registers its listeners.

### 6. Store implementations under `store/` subdirectory

**Decision**: Each backend implementation lives in `pkg/discovery/store/<name>/` (e.g., `store/ram/`, `store/etcd/`).

**Rationale**: Keeps the implementation detail isolated. Users import only the backend they need. The `ram` implementation uses only stdlib (sync primitives, maps) for testing.

## Risks / Trade-offs

- **[Wire format drift]** → The JSON field names and enum values must stay synchronized with the TS implementation. Mitigation: use the same field naming convention (`json:"..."` tags match TS property names exactly) and the same state enum integer values.
- **[Full-snapshot watch scalability]** → Pushing full snapshots on every change could be expensive with thousands of entities. Mitigation: acceptable for expected cluster size. Can add incremental watch later without breaking the interface.
- **[No node registration in this change]** → Services will register but nodes will not. A node's services will be discoverable but the node itself won't appear in node listings. Mitigation: node registration is explicitly deferred; the interfaces support it already.
- **[Backend implementations deferred]** → Only `store/ram` will ship in this change. Production readiness requires etcd/zk implementations. Mitigation: interfaces are designed for it; `store/ram` validates the architecture end-to-end.
