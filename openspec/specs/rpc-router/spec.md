## ADDED Requirements

### 需求:Router 提供 Method 级路由
Router 必须根据 `pkt.Method` 将 Request 分发到已注册的强类型 handler。

#### 场景:已注册的 Method 收到请求
- **当** Router 收到 `pkt.Method` 为已注册 method 的 Request Packet
- **那么** Router 必须 decode payload 为注册时的 `Req` 类型，调用对应 handler

#### 场景:未注册的 Method 收到请求
- **当** Router 收到 `pkt.Method` 未在 `methodTable` 中注册的 Request Packet
- **那么** Router 必须返回 `ERR_METHOD_NOT_FOUND` 错误 Response，payload 格式为 `{error: {code: "ERR_METHOD_NOT_FOUND", ...}, result: null}`

### 需求:Method 泛型注册
Router 必须提供 `Method[Req, Resp](method, handler)` 方法，handler 签名为 `func(conn *Connection, req Req) (Resp, error)`。

#### 场景:注册 Request handler
- **当** 调用 `Method[EchoReq, EchoResp]("echo", handler)`
- **那么** Router 在 `methodTable` 中注册该 handler，后续 `pkt.Method == "echo"` 的 Request 使用此 handler

#### 场景:Handler 正常返回
- **当** handler 返回 `(resp, nil)`
- **那么** Router 必须发送 Response Packet，payload 为 `{error: null, result: <encoded resp>}`

#### 场景:Handler 返回 errorx.Error
- **当** handler 返回 `(_, *errorx.Error)`
- **那么** Router 必须发送 Response Packet，payload 为 `{error: {code, message, level, name, args 映射自 errorx.Error}, result: null}`

#### 场景:Handler 返回普通 error
- **当** handler 返回 `(_, error)` 且不是 `*errorx.Error`
- **那么** Router 必须发送 Response Packet，payload 为 `{error: {code: "ERR_INTERNAL", level: UNEXPECTED, name: "InternalError", message: err.Error()}, result: null}`

### 需求:Notify 泛型注册
Router 必须提供 `Notify[Msg](method, handler)` 方法，handler 签名为 `func(conn *Connection, msg Msg) error`。

#### 场景:注册 Notify handler
- **当** 调用 `Notify[ChatMsg]("chat", handler)`
- **那么** Router 在 `notifyTable` 中注册该 handler，后续 `pkt.Method == "chat"` 的 Notify 使用此 handler

#### 场景:Notify handler 返回 error
- **当** Notify handler 返回 error
- **那么** Router 必须仅记录日志，禁止发送任何 Response

#### 场景:未注册的 Method 收到 Notify
- **当** Router 收到 `pkt.Method` 未注册的 Notify Packet
- **那么** Router 必须记录 Warn 级别日志，禁止发送任何 Response

### 需求:Router 输出回调函数
Router 必须提供 `OnRequestCB()` 和 `OnNotifyCB()` 方法，返回值签名与 `ListenerCallbacks.OnRequest` / `OnNotify` 一致。

#### 场景:Listener 侧接线
- **当** 用户构造 `ListenerCallbacks{OnRequest: router.OnRequestCB(), OnNotify: router.OnNotifyCB()}`
- **那么** Listener 收到的请求必须经过 Router 路由分发

#### 场景:Connector 侧接线
- **当** 用户设置 `conn.OnRequest = router.OnRequestCB()` 和 `conn.OnNotify = router.OnNotifyCB()`
- **那么** Connector 收到的请求必须经过 Router 路由分发

### 需求:Middleware 链
Router 必须支持 `Use(middleware)` 注册中间件，middleware 签名为 `func(next DispatchFunc) DispatchFunc`，其中 `DispatchFunc` 签名为 `func(conn *Connection, pkt packet.Packet) error`。

#### 场景:多个 Middleware 执行顺序
- **当** 注册 `mw1`、`mw2`、`mw3` 三个 middleware
- **那么** 请求处理顺序必须为 `mw1 → mw2 → mw3 → 路由分发`

#### 场景:Middleware 中断链
- **当** 某个 middleware 不调用 `next` 直接返回 error
- **那么** 后续 middleware 和路由分发禁止执行

### 需求:Response 使用 Packet Codec
Router 构造 Response 时必须从入站 `Packet.codec` 获取 PayloadCodec，禁止 Router 自身持有 codec。

#### 场景:Response 编码
- **当** Router 需要发送 Response
- **那么** 必须使用 `pkt.codec.Marshal()` 编码 payload，使用 `pkt.codec` 作为 `NewDecodedPacket` 的 codec 参数

### 需求:Decode 失败处理
当 `packet.Decode[Req](pkt)` 失败时，Router 必须发送 `ERR_DECODE_FAILED` 错误 Response。

#### 场景:Request decode 失败
- **当** `packet.Decode[Req](pkt)` 返回 error
- **那么** Router 必须发送 Response，payload 为 `{error: {code: "ERR_DECODE_FAILED", ...}, result: null}`

#### 场景:Notify decode 失败
- **当** `packet.Decode[Msg](pkt)` 返回 error
- **那么** Router 必须仅记录错误日志，禁止发送任何 Response
