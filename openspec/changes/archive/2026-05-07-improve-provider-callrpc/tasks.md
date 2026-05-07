## 1. packet 包新增类型

- [x] 1.1 在 `pkg/rpc/packet/packet.go` 中新增 `PayloadError` struct（Code, Message, Level, Name, Args 字段，含 json 标签）
- [x] 1.2 在 `pkg/rpc/packet/packet.go` 中新增 `Response[T any]` struct（Error *PayloadError, Result T 字段，含 json 标签）

## 2. router 包迁移引用

- [x] 2.1 修改 `pkg/rpc/router/router.go`，删除本地 `PayloadError` 和 `ResPayload` 定义，改为引用 `packet.PayloadError` 和 `packet.Response`
- [x] 2.2 验证 router 包编译通过，所有 `sendSuccessResponse`、`sendErrorResponse`、`sendMethodNotFound` 等函数使用新类型

## 3. Provider 错误统一化

- [x] 3.1 修改 `pkg/rpc/provider/errors.go`，将所有预定义错误的 `Name` 从 `"ProviderError"` 改为 `"RpcError"`

## 4. CallRpc 泛型改造

- [x] 4.1 修改 `pkg/rpc/provider/interface.go`，将 `CallRpc` 签名改为 `CallRpc[Resp any](ctx context.Context, method string, req any, opts ...CallOption) (Resp, error)`，新增 `SendNotify` 方法
- [x] 4.2 修改 `pkg/rpc/provider/provider.go`，重写 `CallRpc` 实现：调用 `callRpcRaw` → `Decode[packet.Response[Resp]]` → 检查 Error 转为 `errorx.Error{Name: "RpcResponseError"}` → 返回 Result
- [x] 4.3 确保 `CallRpc` 中传输层错误（callRpcRaw 返回的 error）直接透传，不包装

## 5. SendNotify 实现

- [x] 5.1 在 `pkg/rpc/provider/rpc_sender.go` 中新增 `sendNotifyRaw(ctx context.Context, method string, payload []byte, headers map[string]string) error` 方法
- [x] 5.2 在 `pkg/rpc/provider/` 中新增 `notify_options.go`，定义 `NotifyOption` 类型和 `WithNotifyTarget` 函数
- [x] 5.3 在 `pkg/rpc/provider/provider.go` 中实现 `SendNotify` 方法：选 sender → marshal → sendNotifyRaw

## 6. 适配与验证

- [x] 6.1 更新 `cmd/sora-test/main.go`，适配新 `CallRpc[EchoResponse]` 签名
- [x] 6.2 全项目编译验证（`go build ./...`）
- [x] 6.3 运行测试验证（`go test ./...`）
