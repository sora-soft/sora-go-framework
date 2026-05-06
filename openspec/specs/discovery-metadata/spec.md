### Requirement: NodeMeta structure
The `NodeMeta` struct SHALL contain the following fields with JSON tags matching the TS `INodeMetaData` interface: `ID` (`json:"id"`), `Alias` (`json:"alias,omitempty"` as `*string`), `Host` (`json:"host"`), `Pid` (`json:"pid"`), `State` (`json:"state"` as `int`), `StartTime` (`json:"startTime"` as `int64`), `Versions` (`json:"versions"`) containing `Framework` and `App` strings.

#### Scenario: NodeMeta serializes to TS-compatible JSON
- **WHEN** a `NodeMeta` is marshaled to JSON
- **THEN** the output SHALL be parseable by the TS `INodeMetaData` interface with matching field names

### Requirement: ServiceMeta structure
The `ServiceMeta` struct SHALL contain: `Name` (`json:"name"`), `ID` (`json:"id"`), `Alias` (`json:"alias,omitempty"` as `*string`), `State` (`json:"state"` as `int`), `NodeID` (`json:"nodeId"`), `StartTime` (`json:"startTime"` as `int64`), `Labels` (`json:"labels"` as `map[string]string`).

#### Scenario: ServiceMeta serializes to TS-compatible JSON
- **WHEN** a `ServiceMeta` is marshaled to JSON
- **THEN** the output SHALL be parseable by the TS `IServiceMetaData` interface with matching field names

### Requirement: WorkerMeta structure
The `WorkerMeta` struct SHALL contain: `Name` (`json:"name"`), `ID` (`json:"id"`), `Alias` (`json:"alias,omitempty"` as `*string`), `State` (`json:"state"` as `int`), `NodeID` (`json:"nodeId"`), `StartTime` (`json:"startTime"` as `int64`).

#### Scenario: WorkerMeta serializes to TS-compatible JSON
- **WHEN** a `WorkerMeta` is marshaled to JSON
- **THEN** the output SHALL be parseable by the TS `IWorkerMetaData` interface with matching field names

### Requirement: EndpointMeta structure
The `EndpointMeta` struct SHALL contain: `ID` (`json:"id"`), `Protocol` (`json:"protocol"`), `Endpoint` (`json:"endpoint"`), `State` (`json:"state"` as `int`), `Labels` (`json:"labels"` as `map[string]string`), `Codecs` (`json:"codecs"` as `[]string`), `Weight` (`json:"weight"` as `int`), `TargetID` (`json:"targetId"`), `TargetName` (`json:"targetName"`).

#### Scenario: EndpointMeta serializes to TS-compatible JSON
- **WHEN** an `EndpointMeta` is marshaled to JSON
- **THEN** the output SHALL be parseable by the TS `IListenerMetaData` interface with matching field names

### Requirement: State enum value compatibility
The `State` field in all metadata types SHALL use integer values identical to the TS `WorkerState` and `ListenerState` enums: Init=1, Pending/Starting=2, Ready=3, Stopping=4, Stopped=5, Error=100.

#### Scenario: State values match TS enum
- **WHEN** a service is in Ready state
- **THEN** the `State` field SHALL serialize to JSON as `3`, matching TS `WorkerState.Ready`

### Requirement: EndpointMeta conversion from ListenerMetaInfo
The discovery package SHALL provide a conversion function that creates an `EndpointMeta` from an `rpc.ListenerMetaInfo`, populating `Protocol`, `Endpoint`, `State`, `Labels`, and `Codecs` from the source. The caller is responsible for setting `ID`, `Weight`, `TargetID`, and `TargetName` after conversion.

#### Scenario: Conversion preserves listener fields
- **WHEN** `NewEndpointMetaFromListener(listenerMeta)` is called with a `ListenerMetaInfo` having Protocol="tcp", Endpoint="0.0.0.0:8080"
- **THEN** the resulting `EndpointMeta` SHALL have Protocol="tcp", Endpoint="0.0.0.0:8080" and ID/Weight/TargetID/TargetName at zero values
