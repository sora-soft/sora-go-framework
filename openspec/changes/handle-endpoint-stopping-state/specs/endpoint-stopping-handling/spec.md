## ADDED Requirements

### 需求:watchLoop 必须在 endpoint 处于 Stopping 状态时销毁对应 RpcSender

当 watchLoop 收到的 snapshot 中某个已连接的 endpoint 其 State 等于 ListenerStateStopping 时，MUST 调用 `removeSenderLocked` 销毁该 endpoint 对应的 RpcSender（触发 Destroy → cancel context → Disconnect）。

#### 场景:已连接的 endpoint 进入 Stopping 状态
- **当** watchLoop 收到 snapshot，其中已存在 sender 的 endpoint 的 State 变为 Stopping
- **那么** 该 endpoint 对应的 RpcSender MUST 被 Destroy，连接断开且不重连

#### 场景:多个 endpoint 中部分进入 Stopping 状态
- **当** watchLoop 收到 snapshot，其中 3 个 endpoint 有 1 个 State 为 Stopping
- **那么** 只有 Stopping 的 endpoint 对应的 sender 被 Destroy，其余 sender 不受影响

### 需求:watchLoop 必须阻止为 Stopping 状态的 endpoint 创建新 RpcSender

当 watchLoop 收到的 snapshot 中某个 endpoint 不在当前 sender 列表中（新发现），但 State 等于 ListenerStateStopping 时，MUST 跳过该 endpoint，不调用 `addSenderLocked`。

#### 场景:新发现的 endpoint 已经在 Stopping 状态
- **当** watchLoop 收到 snapshot，其中包含一个当前 sender 列表中不存在的新 endpoint，且其 State 为 Stopping
- **那么** MUST 不为该 endpoint 创建 RpcSender

#### 场景:Stopping 的 endpoint 消失后重新出现为 Ready 状态
- **当** 一个之前被跳过的 endpoint 在后续 snapshot 中重新出现且 State 不再为 Stopping
- **那么** 该 endpoint MUST 被正常创建 RpcSender
