## ADDED Requirements

### Requirement: Runtime SHALL expose FrameLogger with identify "framework"
Runtime SHALL create a `*logger.Logger` with identify `"framework"` at construction time and expose it as the exported field `FrameLogger`.

#### Scenario: FrameLogger has correct identify
- **WHEN** `NewRuntime()` is called
- **THEN** `rt.FrameLogger` SHALL be a non-nil `*logger.Logger` with identify `"framework"`

### Requirement: Runtime SHALL expose RpcLogger with identify "rpc"
Runtime SHALL create a `*logger.Logger` with identify `"rpc"` at construction time and expose it as the exported field `RpcLogger`.

#### Scenario: RpcLogger has correct identify
- **WHEN** `NewRuntime()` is called
- **THEN** `rt.RpcLogger` SHALL be a non-nil `*logger.Logger` with identify `"rpc"`

### Requirement: Users SHALL configure Logger outputs directly
Users SHALL call `rt.FrameLogger.AddOutput(...)` and `rt.RpcLogger.AddOutput(...)` to configure outputs. Runtime SHALL NOT add any default outputs.

#### Scenario: No default outputs
- **WHEN** `NewRuntime()` is called
- **THEN** `rt.FrameLogger` and `rt.RpcLogger` SHALL have zero outputs until user calls AddOutput
