## ADDED Requirements

### 需求:SendNotify 方法
Provider interface 必须包含 `SendNotify(ctx context.Context, method string, req any, opts ...NotifyOption) error` 方法，用于发送单向通知（不等响应）。

#### 场景:成功发送通知
- **当** 调用 `p.SendNotify(ctx, "userCreated", msg)` 且存在可用 sender
- **那么** 系统选择 sender、序列化 msg、构建 PacketOpcodeNotify 的 Packet 并发送，返回 nil

#### 场景:无可用 sender
- **当** 调用 `SendNotify` 但没有可用的 sender
- **那么** 返回 `ErrNoAvailableEndpoint` 错误（`*errorx.Error{Name: "RpcError"}`）

#### 场景:Provider 已停止
- **当** 调用 `SendNotify` 但 Provider 的 context 已取消
- **那么** 返回 `ErrProviderStopped` 错误（`*errorx.Error{Name: "RpcError"}`）

### 需求:NotifyOption
系统必须在 `pkg/rpc/provider` 包中定义 `NotifyOption` 类型和 `WithNotifyTarget(targetID string) NotifyOption` 选项函数，用于指定通知的目标 service。

#### 场景:指定目标发送
- **当** 调用 `p.SendNotify(ctx, "userCreated", msg, WithNotifyTarget("svc-1"))`
- **那么** 系统通过 `selectSenderByService("svc-1")` 选择 sender 发送通知

### 需求:RpcSender.sendNotifyRaw
RpcSender 必须暴露 `sendNotifyRaw(ctx context.Context, method string, payload []byte, headers map[string]string) error` 方法。该方法构建 `PacketOpcodeNotify` 的 Packet，通过 conn.SendRaw 发送，不注册 pending、不等待响应。

#### 场景:发送 Notify Packet
- **当** 调用 `sender.sendNotifyRaw(ctx, "event", payload, headers)`
- **那么** 构建 Opcode 为 Notify 的 Packet，通过 conn.SendRaw 发送，直接返回发送结果

#### 场景:连接不存在
- **当** 调用 `sendNotifyRaw` 但 conn 为 nil
- **那么** 返回 `ErrConnectionLost` 错误
