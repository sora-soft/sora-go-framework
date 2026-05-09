## MODIFIED Requirements

### 需求:Method 泛型注册
Router 必须提供 `Method[Req, Resp](method, handler)` 方法，handler 签名为 `func(ctx *RequestContext, req Req) (Resp, error)`。

#### 场景:注册 Request handler
- **当** 调用 `Method[EchoReq, EchoResp]("echo", handler)`
- **那么** Router 在 `methodTable` 中注册该 handler，后续 `pkt.Method == "echo"` 的 Request 使用此 handler

#### 场景:Handler 正常返回
- **当** handler 返回 `(resp, nil)`
- **那么** Router 必须发送 Response Packet，payload 为 `{error: null, result: <encoded resp>}`，Headers 为请求 Headers 与 `ctx.res.headers` 的合并结果

#### 场景:Handler 返回 errorx.Error
- **当** handler 返回 `(_, *errorx.Error)`
- **那么** Router 必须发送 Response Packet，payload 为 `{error: {code, message, level, name, args 映射自 errorx.Error}, result: null}`，Headers 为请求 Headers 与 `ctx.res.headers` 的合并结果

#### 场景:Handler 返回普通 error
- **当** handler 返回 `(_, error)` 且不是 `*errorx.Error`
- **那么** Router 必须发送 Response Packet，payload 为 `{error: {code: "ERR_INTERNAL", level: UNEXPECTED, name: "InternalError", message: err.Error()}, result: null}`，Headers 为请求 Headers 与 `ctx.res.headers` 的合并结果

### 需求:Notify 泛型注册
Router 必须提供 `Notify[Msg](method, handler)` 方法，handler 签名为 `func(ctx *NotifyContext, msg Msg) error`。

#### 场景:注册 Notify handler
- **当** 调用 `Notify[ChatMsg]("chat", handler)`
- **那么** Router 在 `notifyTable` 中注册该 handler，后续 `pkt.Method == "chat"` 的 Notify 使用此 handler

#### 场景:Notify handler 返回 error
- **当** Notify handler 返回 error
- **那么** Router 必须仅记录日志，禁止发送任何 Response

#### 场景:未注册的 Method 收到 Notify
- **当** Router 收到 `pkt.Method` 未注册的 Notify Packet
- **那么** Router 必须记录 Warn 级别日志，禁止发送任何 Response

### 需求:Middleware 链
Router 必须支持 `Use(middleware)` 注册中间件，middleware 签名为 `func(next DispatchFunc) DispatchFunc`，其中 `DispatchFunc` 签名为 `func(ctx HandlerContext) error`。

#### 场景:多个 Middleware 执行顺序
- **当** 注册 `mw1`、`mw2`、`mw3` 三个 middleware
- **那么** 请求处理顺序必须为 `mw1 → mw2 → mw3 → 路由分发`

#### 场景:Middleware 中断链
- **当** 某个 middleware 不调用 `next` 直接返回 error
- **那么** 后续 middleware 和路由分发禁止执行

#### 场景:Middleware 读取请求头
- **当** middleware 调用 `ctx.Reader().Header("authorization")`
- **那么** 必须返回请求 Packet 中对应的 Header 值

#### 场景:Middleware 向 handler 传值
- **当** middleware 调用 `ctx.Set("user", userObj)`
- **那么** 后续 middleware 和 handler 通过 `ctx.Get("user")` 必须获得 `userObj`

#### 场景:Middleware 通过断言写入响应头
- **当** middleware 需要设置响应头，执行 `rc, ok := ctx.(*RequestContext)`，`ok == true` 时调用 `rc.Res().SetHeader("x-foo", "bar")`
- **那么** 响应 Headers 必须包含 `"x-foo": "bar"`
