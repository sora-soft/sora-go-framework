## ADDED Requirements

### Requirement: Listener 状态机
Listener SHALL 实现状态机：`Init(1) → Starting(2) → Ready(3) → Stopping(4) → Stopped(5)`，任意状态可进入 `Error(100)`。状态迁移 SHALL 使用 `utility.LifeCycle[T]`。

#### Scenario: 正常启动流程
- **WHEN** 调用 `Start(ctx)` 且 TransportListener 已绑定
- **THEN** 状态依次为 Init → Starting → Ready，Start 阻塞直到 Ready 后返回 nil

#### Scenario: 启动失败
- **WHEN** 调用 `Start(ctx)` 且启动过程中发生错误
- **THEN** 状态为 Error，Start 返回错误

#### Scenario: 正常关闭
- **WHEN** 调用 `Stop()` 且当前状态为 Ready
- **THEN** 状态依次为 Ready → Stopping → Stopped，所有 session 被关闭

### Requirement: Listener.Start 阻塞
Listener 的 `Start` 方法 SHALL 阻塞直到状态变为 Ready 或发生错误。Start 返回后 Listener SHALL 已就绪并可接受连接。

#### Scenario: Start 阻塞直到就绪
- **WHEN** 调用 `Start(ctx)`
- **THEN** Start 阻塞，直到 Listener 进入 Ready 状态后返回 nil

#### Scenario: Start 期间 context 取消
- **WHEN** 调用 `Start(ctx)` 且 ctx 被取消
- **THEN** Start 返回 context 错误

### Requirement: Listener.Stop 关闭所有 session
Listener 的 `Stop` 方法 SHALL 取消内部 context（终止 acceptLoop）、关闭 TransportListener、对所有已注册 session 调用 `Disconnect()`。

#### Scenario: Stop 关闭所有连接
- **WHEN** 调用 `Stop()` 且存在 3 个活跃 session
- **THEN** 3 个 session 均被 Disconnect，状态变为 Stopped

### Requirement: Listener.CloseSession 关闭单个 session
Listener SHALL 提供 `CloseSession(sessionId string) error` 方法，关闭指定 session 并从 sessions map 中移除。

#### Scenario: 关闭指定 session
- **WHEN** 调用 `CloseSession("abc")` 且该 session 存在
- **THEN** 该 session 被 Disconnect，从 map 中移除，返回 nil

#### Scenario: session 不存在
- **WHEN** 调用 `CloseSession("xyz")` 且该 session 不存在
- **THEN** 返回错误
