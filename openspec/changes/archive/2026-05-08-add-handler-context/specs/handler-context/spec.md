## ADDED Requirements

### 需求:HandlerContext 接口定义
框架必须定义 `HandlerContext` 接口，作为 Request 和 Notify handler 共享的上下文抽象。接口必须包含以下方法：`Conn() *Connection`、`Context() context.Context`、`Reader() *RequestReader`、`Set(k string, v any)`、`Get(k string) any`。

#### 场景:RequestContext 实现 HandlerContext
- **当** 创建 `RequestContext` 实例
- **那么** 该实例必须满足 `HandlerContext` 接口，可通过 `HandlerContext` 类型的参数传递

#### 场景:NotifyContext 实现 HandlerContext
- **当** 创建 `NotifyContext` 实例
- **那么** 该实例必须满足 `HandlerContext` 接口，可通过 `HandlerContext` 类型的参数传递

### 需求:RequestContext 结构体
框架必须提供 `RequestContext` 结构体，用于 Method handler。`RequestContext` 必须通过内嵌 `baseContext` 共享通用实现，并额外提供 `res *ResponseWriter` 字段和 `Res() *ResponseWriter` 方法。

#### 场景:RequestContext 提供完整上下文
- **当** Method handler 接收到 `*RequestContext`
- **那么** handler 必须能够通过 `ctx.Conn()` 获取连接、`ctx.Reader()` 读取请求元数据、`ctx.Res()` 写入响应头、`ctx.Set()/Get()` 传递数据

#### 场景:RequestContext 持有独立的 context.Context
- **当** 创建 `RequestContext` 时
- **那么** `ctx.Context()` 必须初始返回 `context.Background()`，可通过 `WithContext()` 替换

### 需求:NotifyContext 结构体
框架必须提供 `NotifyContext` 结构体，用于 Notify handler。`NotifyContext` 必须通过内嵌 `baseContext` 共享通用实现，不提供 `Res()` 方法。

#### 场景:NotifyContext 无响应写入能力
- **当** Notify handler 接收到 `*NotifyContext`
- **那么** handler 必须能够通过 `ctx.Conn()` 获取连接、`ctx.Reader()` 读取请求元数据、`ctx.Set()/Get()` 传递数据，但禁止通过 `ctx` 写入响应头

### 需求:RequestReader 只读访问
框架必须提供 `RequestReader` 结构体，持有 `packet.Packet` 引用，提供以下只读方法：`Header(key string) string`、`Headers() map[string]string`、`Method() string`、`Service() string`。

#### 场景:读取请求头
- **当** 调用 `ctx.Reader().Header("authorization")`
- **那么** 必须返回请求 Packet 中 `Headers["authorization"]` 的值，不存在时返回空字符串

#### 场景:读取所有请求头
- **当** 调用 `ctx.Reader().Headers()`
- **那么** 必须返回请求 Packet 的完整 `Headers` map

#### 场景:读取 Method 名称
- **当** 调用 `ctx.Reader().Method()`
- **那么** 必须返回请求 Packet 的 `Method` 字段值

#### 场景:读取 Service 名称
- **当** 调用 `ctx.Reader().Service()`
- **那么** 必须返回请求 Packet 的 `Service` 字段值

### 需求:ResponseWriter 响应头写入
框架必须提供 `ResponseWriter` 结构体，持有独立的 `headers map[string]string`，提供以下方法：`SetHeader(key string, value string)`、`SetHeaders(headers map[string]string)`。

#### 场景:设置单个响应头
- **当** 调用 `ctx.Res().SetHeader("x-trace-id", "abc123")`
- **那么** 该 key-value 对必须被记录到响应头 map 中

#### 场景:设置多个响应头
- **当** 调用 `ctx.Res().SetHeaders(map[string]string{"k1": "v1", "k2": "v2"})`
- **那么** 所有 key-value 对必须被合并到响应头 map 中

#### 场景:覆盖同名响应头
- **当** 先调用 `SetHeader("x-foo", "a")` 再调用 `SetHeader("x-foo", "b")`
- **那么** 最终 `headers["x-foo"]` 必须为 `"b"`

### 需求:Middleware→Handler 数据传递
`baseContext` 必须提供 `Set(k string, v any)` 和 `Get(k string) any` 方法，支持 middleware 向 handler 传递数据。

#### 场景:Middleware 设置数据，Handler 读取
- **当** middleware 调用 `ctx.Set("user", userObj)`，handler 调用 `ctx.Get("user")`
- **那么** handler 必须获得 `userObj`

#### 场景:Key 不存在
- **当** 调用 `ctx.Get("nonexistent")`
- **那么** 必须返回 `nil`

### 需求:Middleware 通过类型断言区分流
Middleware 必须能够通过类型断言 `ctx.(*RequestContext)` 区分 Request 流和 Notify 流。

#### 场景:Request 流中断言成功
- **当** Request 流中的 middleware 执行 `rc, ok := ctx.(*RequestContext)`
- **那么** `ok` 必须为 `true`，`rc.Res()` 必须返回有效的 `*ResponseWriter`

#### 场景:Notify 流中断言失败
- **当** Notify 流中的 middleware 执行 `_, ok := ctx.(*RequestContext)`
- **那么** `ok` 必须为 `false`

### 需需:响应头合并
Router 发送 Response 时，必须将请求 Headers 与 handler 通过 ResponseWriter 设置的响应头合并。同名 key 以 ResponseWriter 设置的值为准。

#### 场景:Handler 设置额外的响应头
- **当** 请求 Headers 为 `{"x-sora-rpc-id": "1"}`，handler 调用 `ctx.Res().SetHeader("x-custom", "value")`
- **那么** 响应 Packet 的 Headers 必须为 `{"x-sora-rpc-id": "1", "x-custom": "value"}`

#### 场景:Handler 覆盖请求头
- **当** 请求 Headers 为 `{"x-sora-rpc-id": "1", "x-foo": "old"}`，handler 调用 `ctx.Res().SetHeader("x-foo", "new")`
- **那么** 响应 Packet 的 Headers 必须为 `{"x-sora-rpc-id": "1", "x-foo": "new"}`
