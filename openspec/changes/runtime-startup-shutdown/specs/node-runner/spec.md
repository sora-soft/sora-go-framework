## MODIFIED Requirements

### Requirement: NodeRunner installs Listeners during Startup
NodeRunner SHALL accept a slice of `*rpc.Listener` at construction. During `Startup`, NodeRunner SHALL iterate over the listeners and call `svcRef.InstallListener(ctx, l)` for each one. NodeRunner SHALL access the global Runtime as `RT` (same package, no import required).

#### Scenario: Startup installs all pre-configured listeners
- **WHEN** `Startup(ctx)` is called on NodeRunner with 3 listeners configured
- **THEN** `svcRef.InstallListener` SHALL be called exactly 3 times, once per listener

#### Scenario: Listener install failure aborts Startup
- **WHEN** `svcRef.InstallListener` returns an error for any listener
- **THEN** Startup SHALL return that error immediately without installing remaining listeners

### Requirement: NodeRunner registers NodeId during Startup
NodeRunner SHALL call `RT.SetNodeId(svc.GetMetadata().Id)` during Startup after installing listeners. The NodeId SHALL be the ID of the wrapping Service. `RT` SHALL be accessed directly without package qualifier.

#### Scenario: Startup sets NodeId on Runtime
- **WHEN** `Startup(ctx)` completes successfully
- **THEN** `RT.NodeId()` SHALL return the Service's ID
