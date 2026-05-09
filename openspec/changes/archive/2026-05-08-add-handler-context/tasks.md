## 1. 新增核心类型

- [x] 1.1 创建 `pkg/rpc/reader.go`：实现 `RequestReader` 结构体（持有 `packet.Packet`），提供 `Header(key) string`、`Headers() map[string]string`、`Method() string`、`Service() string` 方法
- [x] 1.2 创建 `pkg/rpc/writer.go`：实现 `ResponseWriter` 结构体（持有 `headers map[string]string`），提供 `SetHeader(k, v string)`、`SetHeaders(headers map[string]string)` 方法
- [x] 1.3 创建 `pkg/rpc/context.go`：定义 `HandlerContext` 接口（`Conn()`、`Context()`、`Reader()`、`Set()`、`Get()`），实现 `baseContext` 内嵌结构体、`RequestContext`（含 `Res()`）和 `NotifyContext`
- [x] 1.4 为 `RequestContext` 添加 `WithContext(ctx context.Context) *RequestContext` 方法，为 `NotifyContext` 添加 `WithContext(ctx context.Context) *NotifyContext` 方法

## 2. 重构 Router 签名与 dispatch

- [x] 2.1 修改 `DispatchFunc` 签名为 `func(ctx HandlerContext) error`，修改 `Middleware` 类型定义
- [x] 2.2 重构 `OnRequestCB`：创建 `RequestContext` 实例，传入 middleware chain
- [x] 2.3 重构 `OnNotifyCB`：创建 `NotifyContext` 实例，传入 middleware chain
- [x] 2.4 重构 `dispatchRequest`：接受 `HandlerContext`，类型断言为 `*RequestContext` 后分发
- [x] 2.5 重构 `dispatchNotify`：接受 `HandlerContext`，类型断言为 `*NotifyContext` 后分发
- [x] 2.6 重构 `Method[Req, Resp]` 泛型注册：handler 签名改为 `func(ctx *RequestContext, req Req) (Resp, error)`，内部包装器类型断言 `*RequestContext`
- [x] 2.7 重构 `Notify[Msg]` 泛型注册：handler 签名改为 `func(ctx *NotifyContext, msg Msg) error`，内部包装器类型断言 `*NotifyContext`

## 3. 响应头合并

- [x] 3.1 修改 `sendSuccessResponse`：合并请求 Headers 与 `ctx.res.headers`，handler 设置的同名 key 覆盖请求头
- [x] 3.2 修改 `sendErrorResponse`、`sendMethodNotFoundResponse`、`sendDecodeErrorResponse`：同样合并响应头（从 panic recovery 上下文中提取，若无 RequestContext 则回退为仅请求头）

## 4. 更新示例

- [x] 4.1 更新 `cmd/sora-test/main.go`：将所有 handler 签名从 `(conn *rpc.Connection, ...)` 改为对应的 `(ctx *rpc.RequestContext, ...)` / `(ctx *rpc.NotifyContext, ...)`

## 5. 验证

- [x] 5.1 编译通过：`go build ./...`
- [x] 5.2 运行 `cmd/sora-test` 验证 echo 请求/响应、notify 收发正常工作
