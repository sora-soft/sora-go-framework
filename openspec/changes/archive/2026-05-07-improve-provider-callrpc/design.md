## Context

Provider (`pkg/rpc/provider`) 是 RPC 客户端，通过 `WatchEndpoints` 监听服务发现，管理多个 `RpcSender`，提供 RPC 调用能力。当前 `CallRpc` 返回原始 `packet.Packet`，调用者需要手动反序列化并检查业务错误。

`ResPayload[T]` 和 `PayloadError` 定义在服务端 `router` 包中，但 `{error, result}` 是传输协议格式，客户端也需要理解——这是包归属不合理的问题。

Provider 缺少单向通知（Notify）能力，目前只能发送 Request/Response 模式的调用。

## Goals / Non-Goals

**Goals:**
- `CallRpc` 直接返回反序列化后的业务结果，消除调用侧样板代码
- `ResPayload` / `PayloadError` 迁移到 `packet` 公共包
- 错误统一为 `errorx.Error`，通过 `Name` 区分本地错误（`"RpcError"`）和远端错误（`"RpcResponseError"`）
- 新增 `SendNotify` 方法支持单向通知

**Non-Goals:**
- 不改变 Router / 服务端的行为（仅迁移类型引用）
- 不引入重试、熔断等高级调用策略
- 不改变 `RpcSender` 的连接管理和 pending 机制

## Decisions

### Decision 1: CallRpc 使用泛型签名

**选择**: `CallRpc[Resp any](ctx, method, req, opts...) (Resp, error)`

**替代方案**:
- A: 保持返回 `Packet`，增加 `CallRpcTyped[Resp]` 辅助方法 → 两套 API 增加认知负担
- B: 返回 `any`，调用者自行断言 → 丢失类型安全

**理由**: 泛型方法让调用者一步拿到类型安全的结果，interface 方法支持泛型参数在 Go 中可行。`callRpcRaw` 内部仍返回 `Packet`，泛型包装只在 Provider 层。

### Decision 2: PayloadError / ResPayload 迁移到 packet 包

**选择**: 搬迁到 `packet` 包，重命名为 `PayloadError`（保持原名）和 `Response[T]`

**替代方案**:
- A: 留在 router 包，provider 引用 router → 客户端反向依赖服务端包，不合理
- B: 新建 `protocol` 包 → 过度拆分，packet 包已经承担协议模型职责

**理由**: `Response[T]` 描述的是线路格式（`{error, result}`），属于协议层。packet 包已有的 `Packet`、`PacketOpcode` 等类型也承担协议模型职责，`Response` 放在这里自然。

### Decision 3: 错误统一化

**选择**: 所有错误都是 `*errorx.Error`，通过 `Name` 字段区分

| Name | 含义 | 来源 |
|------|------|------|
| `"RpcError"` | 发送端/本地错误 | Provider / RpcSender 层 |
| `"RpcResponseError"` | 远端业务错误 | 从 `Response.Error` 转换 |

**替代方案**:
- A: 定义 `RpcError` / `RpcResponseError` 两个独立类型 → 增加类型数量，errorx 已有足够表达能力
- B: 只用 `fmt.Errorf` → 丢失结构化信息，无法按 Code 区分

**理由**: `errorx.Error` 已有 `Code`、`Level`、`Name`、`Message`、`Extra` 字段，完全覆盖 `PayloadError` 的所有信息。`Name` 字段天然适合做分类标签，不需要引入新类型。

### Decision 4: SendNotify 设计

**选择**: `SendNotify(ctx, method, req, opts ...NotifyOption) error`

- `NotifyOption` 独立于 `CallOption`，但提供 `WithNotifyTarget(string)` 实现目标指定
- RpcSender 暴露 `sendNotifyRaw(ctx, method, payload, headers)` 方法
- 内部构建 `PacketOpcodeNotify` 的 Packet，直接发送不等响应

**理由**: Notify 是 fire-and-forget，不需要 pending map、requestId、超时机制。但仍然需要选 sender 和 ready 检查，所以走 sender 统一入口。

## Risks / Trade-offs

- **[BREAKING CHANGE]** CallRpc 签名变更 → 所有调用方必须更新。影响范围目前仅 `cmd/sora-test`，可控
- **泛型 interface 方法** → Go 1.18+ 支持，项目已使用泛型，无额外风险
- **PayloadError 字段名** → `PayloadError.Args` 是 `any` 类型，转为 `errorx.Error.Extra` 时无损失
- **SendNotify 无确认** → 调用者无法知道对端是否收到。这是 Notify 语义的固有特性，文档说明即可
