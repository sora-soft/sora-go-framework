## Why

当外部节点关闭时，会通过 Listener lifecycle 将 endpoint 状态更新为 Stopping 并注册到 discovery，同时向已连接的 session 发送 `off` 命令。但当前 Go 端的 RpcSender 通过 WatchEndpoints 收到状态更新后，完全不检查 endpoint 的 State 字段，导致连接不断开，外部进程无法优雅关闭。

## What Changes

- RpcSender 的 `watchLoop` 需要识别 endpoint 的 Stopping 状态，在检测到 Stopping 时主动 `removeSender`（调用 Destroy → cancel context → 断开连接 + 不重连）
- Stopping 状态的 endpoint 不应被 `addSender` 创建新连接

## Capabilities

### New Capabilities

- `endpoint-stopping-handling`: RpcSender 通过 discovery 的 endpoint state 变化，主动断开正在 Stopping 的 endpoint 连接，并不再重连

### Modified Capabilities

## Impact

- `pkg/rpc/provider/provider.go` — watchLoop 逻辑变更，增加 State 检查
- `pkg/discovery/metadata.go` — EndpointMeta.State 字段已有，无需变更
- `pkg/rpc/listener.go` — Stopping 状态注册到 discovery 的逻辑已有，无需变更
