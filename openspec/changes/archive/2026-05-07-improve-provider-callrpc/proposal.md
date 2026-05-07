## Why

Provider 的 `CallRpc` 当前返回 `packet.Packet`，调用者必须手动 `Decode[router.ResPayload[T]]` 再检查 `resp.Error`，每次调用需要 4+ 行样板代码。`ResPayload` 和 `PayloadError` 定义在服务端的 `router` 包中，客户端 Provider 引用服务端类型不合理——`{error, result}` 是传输协议的一部分，应属于公共层。此外 Provider 缺少 `SendNotify` 方法，无法作为客户端发送单向通知。

## What Changes

- **BREAKING**: `CallRpc` 签名从 `(packet.Packet, error)` 改为 `CallRpc[Resp any](...) (Resp, error)`，内部完成反序列化和错误转换，直接返回业务结果
- 将 `router.ResPayload[T]` 和 `router.PayloadError` 搬到 `packet` 包，重命名为 `packet.Response[T]` 和 `packet.PayloadError`
- 新增 `SendNotify(ctx, method, req, opts ...NotifyOption) error` 方法到 Provider interface
- Provider 错误统一使用 `errorx.Error`，通过 `Name` 字段区分：`"RpcError"`（发送端/本地错误）和 `"RpcResponseError"`（远端业务错误）
- 现有 provider errors 的 `Name` 从 `"ProviderError"` 改为 `"RpcError"`

## Capabilities

### New Capabilities
- `provider-notify`: Provider 作为客户端发送单向 Notify 的能力，包含 NotifyOption 和 sender 端 sendNotifyRaw 实现

### Modified Capabilities
- `provider-interface`: CallRpc 签名变为泛型，新增 SendNotify 方法，错误统一化
- `packet-model`: 新增 Response[T] 和 PayloadError 类型（从 router 包迁入）

## Impact

- `pkg/rpc/provider/interface.go` — Provider interface 签名变更
- `pkg/rpc/provider/provider.go` — CallRpc 实现重写
- `pkg/rpc/provider/rpc_sender.go` — 新增 sendNotifyRaw
- `pkg/rpc/provider/errors.go` — Name 字段值从 "ProviderError" 改为 "RpcError"
- `pkg/rpc/packet/packet.go` — 新增 Response[T]、PayloadError
- `pkg/rpc/router/router.go` — 引用 packet.Response / packet.PayloadError，删除本地定义
- `cmd/sora-test/main.go` — 适配新 API
- **BREAKING**: 所有调用 `CallRpc` 的外部代码需要更新
