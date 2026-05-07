## 1. 基础类型定义

- [x] 1.1 创建 `pkg/rpc/router/` 包，定义 `Router` struct（`methodTable`、`notifyTable`、`middlewares`、`logger`）
- [x] 1.2 定义 `DispatchFunc` 类型 `func(conn *Connection, pkt packet.Packet) error`
- [x] 1.3 定义 `Middleware` 类型 `func(next DispatchFunc) DispatchFunc`
- [x] 1.4 定义 `PayloadError` struct（Code、Message、Level、Name、Args），定义 `ResPayload[T]` struct（Error、Result）
- [x] 1.5 定义 `Logger` 接口（或使用标准 `log/slog`），用于 Notify 错误日志和 Warn 日志

## 2. 核心路由实现

- [x] 2.1 实现 `NewRouter(opts ...RouterOption) *Router` 构造函数
- [x] 2.2 实现 `Method[Req, Resp](method, handler)` —— 闭包捕获泛型，包装为 `func(conn, pkt) error` 存入 `methodTable`
- [x] 2.3 实现 `Notify[Msg](method, handler)` —— 闭包捕获泛型，包装为 `func(conn, pkt)` 存入 `notifyTable`
- [x] 2.4 实现 `Use(middleware)` —— 追加 middleware 到 `middlewares` 切片

## 3. Response 封装

- [x] 3.1 实现 `sendSuccessResponse(conn, pkt, resp)` —— 从 `pkt.codec` 取 codec，编码 `{error: null, result: resp}`，构造 Response Packet，调用 `conn.SendResponse`
- [x] 3.2 实现 `sendErrorResponse(conn, pkt, err)` —— type switch 处理 `*errorx.Error`（直接映射）和普通 error（兜底 ERR_INTERNAL），编码 `{error: {...}, result: null}`，构造 Response Packet 发送
- [x] 3.3 实现 `sendMethodNotFoundResponse(conn, pkt, method)` —— 构造 `ERR_METHOD_NOT_FOUND` 错误并发送

## 4. 分发逻辑

- [x] 4.1 实现 `dispatchRequest(conn, pkt) error` —— 查 `methodTable`，未找到则发 ERR_METHOD_NOT_FOUND；找到则调用 entry，内含 Decode → handler → sendResponse/sendError
- [x] 4.2 实现 `dispatchNotify(conn, pkt)` —— 查 `notifyTable`，未找到则记 Warn 日志；找到则调用 entry，内含 Decode → handler，error 仅记日志
- [x] 4.3 实现 middleware 链组装 —— `buildChain(core DispatchFunc) DispatchFunc`，按注册顺序 wrap

## 5. 回调输出

- [x] 5.1 实现 `OnRequestCB()` —— 返回 `func(conn, pkt)`，内部组装 middleware 链 + `dispatchRequest`，含 panic recover
- [x] 5.2 实现 `OnNotifyCB()` —— 返回 `func(conn, pkt)`，内部组装 middleware 链 + `dispatchNotify`，含 panic recover

## 6. 集成验证

- [x] 6.1 重构 `cmd/sora-test/main.go` 使用 Router 注册 echo handler，验证 Listener 侧接线正常
- [x] 6.2 添加 Connector 侧 Router 使用示例（如需要）
- [x] 6.3 验证 Response payload 格式符合 `{error: ..., result: ...}` 结构
- [x] 6.4 验证未注册 method 返回 `ERR_METHOD_NOT_FOUND`
- [x] 6.5 验证 Notify handler 错误仅记日志不发送 Response
