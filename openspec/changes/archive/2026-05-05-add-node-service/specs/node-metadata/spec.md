## ADDED Requirements

### Requirement: NodeMetaData structure
`NodeMetaData` SHALL be a struct containing `Id string`, `Alias *string`, `Host string`, `Pid int`, `Endpoints []rpc.ListenerMetaInfo`, `State WorkerState`, `StartTime int64`, and `Versions NodeVersions`.

#### Scenario: NodeMetaData contains all node identity fields
- **WHEN** a NodeMetaData is constructed
- **THEN** it SHALL include fields for node ID, alias, hostname, process ID, listener endpoints, lifecycle state, start timestamp, and version information

#### Scenario: NodeMetaData JSON serialization
- **WHEN** NodeMetaData is serialized to JSON
- **THEN** fields with zero values SHALL use `omitempty` and not appear in output

### Requirement: NodeVersions structure
`NodeVersions` SHALL be a struct containing `Framework string` and `App string`. Both values SHALL default to "0.0.0".

#### Scenario: NodeVersions defaults
- **WHEN** NodeVersions is constructed without explicit version values
- **THEN** Framework SHALL be "0.0.0" and App SHALL be "0.0.0"

### Requirement: NodeRunData structure
`NodeRunData` SHALL be a struct containing `Node NodeMetaData`, `Components []component.ComponentMetadata`, `Services []WorkerMetaData`, and `Workers []WorkerMetaData`.

#### Scenario: NodeRunData aggregates all process metadata
- **WHEN** RunData() is called
- **THEN** the returned NodeRunData SHALL contain the Node's own metadata plus metadata for all registered components, services, and workers

#### Scenario: NodeRunData with empty registries
- **WHEN** RunData() is called with no components, services, or workers registered
- **THEN** Components, Services, and Workers slices SHALL be empty (not nil)
