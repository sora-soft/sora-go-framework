## MODIFIED Requirements

### 需求:Provider 接口定义
系统必须在 `pkg/rpc/provider` 包中定义 `Provider` interface，包含以下方法：
- `Start(ctx context.Context) error`
- `Stop() error`
- `CallRpc[Resp any](ctx context.Context, method string, req any, opts ...CallOption) (Resp, error)`
- `SendNotify(ctx context.Context, method string, req any, opts ...NotifyOption) error`

#### 场景:interface 满足
- **当** 定义了 Provider interface
- **那么** rpcProvider（具体实现）必须自动满足该 interface

#### 场景:CallRpc 返回业务结果
- **当** 调用 `p.CallRpc[EchoResp](ctx, "echo", req)` 且远端正常响应
- **那么** 返回反序列化后的 `EchoResp` 结果，error 为 nil

#### 场景:CallRpc 远端业务错误
- **当** 调用 `CallRpc` 且远端返回 error payload（如方法不存在）
- **那么** 返回 `*errorx.Error{Name: "RpcResponseError"}`，Code 和 Message 来自远端响应

#### 场景:CallRpc 发送端错误
- **当** 调用 `CallRpc` 但发生传输错误（连接丢失、超时等）
- **那么** 返回 `*errorx.Error{Name: "RpcError"}`，Code 标识具体错误类型

### 需求:Provider 错误名称
Provider 包中所有预定义错误（`ErrNoAvailableEndpoint`、`ErrServiceNotFound`、`ErrCallTimeout`、`ErrConnectionLost`、`ErrSenderStopped`、`ErrProviderStopped`、`ErrNoAvailableCodec`）的 `Name` 字段必须为 `"RpcError"`。

#### 场景:错误 Name 统一
- **当** Provider 返回任何本地错误
- **那么** `errors.As(err, &e)` 后 `e.Name == "RpcError"`
