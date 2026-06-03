## Context

当前 Provider 的 `watchLoop` 通过 `WatchEndpoints` 监听 discovery 中的 endpoint 变化，但只处理三种情况：新增 endpoint（创建 sender）、endpoint 消失（销毁 sender）、endpoint 的 Protocol/Endpoint/Codecs 变化（重建 sender）。完全忽略了 `EndpointMeta.State` 字段。

关闭侧已经实现了正确的行为：Listener 进入 Stopping 状态时，通过 lifecycle 监听器将 `EndpointMeta{State: ListenerStateStopping}` 注册到 discovery。但消费侧不读取这个状态。

当前关闭流程：
```
Service.Stop()
  → ListenerStateStopping → RegisterEndpoint({State: 4}) → discovery
  → SendCommand("off") to sessions
  → TCPListener.Close() → connWg.Wait(10s timeout)
```

消费侧完全无感知，RpcSender 持续保持连接和重连。

## Goals / Non-Goals

**Goals:**
- watchLoop 识别 endpoint State 为 Stopping 时，主动销毁对应 RpcSender
- Stopping 状态的 endpoint 不创建新 RpcSender

**Non-Goals:**
- 不处理 connector 侧的 `off` 命令（后续变更）
- 不修改 Listener 侧的关闭逻辑（已有，无需变更）
- 不引入新的 endpoint 状态或 discovery 协议变更

## Decisions

### 1. watchLoop 中 Stopping endpoint 的处理策略

**决策：** 在现有 add/remove/recreate 逻辑中增加 State 检查分支。

在 `watchLoop` 的 snapshot 处理中，对每个 filtered endpoint：
- 如果 endpoint 是新的（不在 currentIds 中）且 State == Stopping → 跳过，不创建 sender
- 如果 endpoint 已存在且 State == Stopping → 调用 `removeSenderLocked` 销毁 sender
- 其余情况保持现有逻辑不变

**替代方案：** 修改 `endpointChanged` 增加 State 比较。被否决，因为 State 变化不应该触发 "remove + add"（重建连接），而是应该触发 "remove only"（断开且不重连）。

### 2. Stopping 状态值的判断

**决策：** 使用 `rpc.ListenerStateStopping`（值为 4）作为判断标准。`EndpointMeta.State` 存储的是 `int`，来源是 `ListenerMetaInfo.State`，类型为 `ListenerState`。

直接比较 `ep.State == int(rpc.ListenerStateStopping)` 即可。

## Risks / Trade-offs

- **[快照延迟]** WatchEndpoints 是 snapshot 模式，存在延迟。在 discovery 更新传播之前，RpcSender 可能短暂重连到正在关闭的 endpoint。 → 可接受，listener 侧的 `tl.Close()` 会在 10s 超时后强制关闭，且 `off` 命令作为后续变更会作为兜底。
- **[Stopping 与 Stopped 的时序]** 如果 listener 快速从 Stopping 过渡到 Stopped（endpoint 被注销），watchLoop 可能只看到 endpoint 消失而从未看到 Stopping 状态。 → 无影响，endpoint 消失时已有的 removeSender 逻辑会正确处理。
