## 上下文

sora-go-framework 是一个 TCP 层 RPC 框架。当前 Router 的类型安全 handler 通过泛型注册（`Method[Req, Resp]`、`Notify[Msg]`），handler 签名只接收 `*Connection` 和解码后的 body，无法感知请求 Headers、无法设置响应 Headers、无法接收 middleware 传递的数据。

Middleware 链基于 `DispatchFunc func(conn *Connection, pkt packet.Packet) error`，可以访问完整 Packet，但这一能力在到达用户 handler 前被"窄化"了。

当前 Packet 结构中已有 `Headers map[string]string` 字段，wire protocol（JSON codec）已支持 Headers 序列化/反序列化。响应包当前直接回显请求 Headers。

## 目标 / 非目标

**目标：**
- 为 handler 提供请求 Headers 的只读访问
- 为 handler 提供响应 Headers 的写入能力
- 为 middleware→handler 提供数据传递通道（key-value store）
- Middleware 链与 handler 使用统一的上下文接口
- 保持 Request（有响应）和 Notify（无响应）的类型区分

**非目标：**
- 不修改 Packet 结构或 wire protocol
- 不修改客户端侧（Provider/CallRpc）
- 不提供 streaming/RPC streaming 支持
- 不实现 HTTP 风格的 status code 或 content negotiation

## 决策

### 决策 1：接口 + 嵌入的类型体系

选择定义 `HandlerContext` 接口作为统一入口，`RequestContext` 和 `NotifyContext` 通过嵌入 `baseContext` 共享实现。

```
HandlerContext (interface)
  ├── Conn()    *Connection
  ├── Context() context.Context
  ├── Reader()  *RequestReader
  ├── Set(k string, v any)
  ├── Get(k string) any

baseContext (unexported)
  ├── conn, ctx, reader, store

RequestContext  { baseContext + res *ResponseWriter }
  └── Res() *ResponseWriter

NotifyContext   { baseContext }
```

**理由**：Go 的接口+嵌入组合提供了自然的多态。Middleware 接受 `HandlerContext`，无需关心具体类型。需要写响应头的 middleware 通过类型断言 `ctx.(*RequestContext)` 区分 Request/Notify 流。

**考虑的替代方案**：
- 在 HandlerContext 接口上直接暴露 `Res()` 方法（NotifyContext 返回 nil）——写法简洁但语义不清晰，违反"Notify 没有 response"的设计意图
- 两条独立的 middleware 链——类型最安全但增加框架复杂度和使用者心智负担

### 决策 2：RequestReader / ResponseWriter 分离读职责

`RequestReader` 持有 `packet.Packet` 引用，提供 `Header()`、`Headers()`、`Method()`、`Service()` 等只读方法。`ResponseWriter` 持有独立的 `headers map[string]string`，提供 `SetHeader()`、`SetHeaders()` 写入方法。

**理由**：读写职责分离让 API 意图明确。`RequestReader` 是不可变的，`ResponseWriter` 是可变的，用户一眼就能看出哪些操作有副作用。

### 决策 3：context.Context 放在 RequestContext 上

通过 `Context()` 和 `WithContext()` 方法访问/修改 Go 标准库的 `context.Context`。

**理由**：`context.Context` 是请求生命周期级别的概念，不属于"读请求"或"写响应"的范畴，应放在外层 Context 上。

### 决策 4：响应头合并策略

Handler 通过 `ResponseWriter` 设置的响应头与原始请求 Headers 合并后写入响应包。同名 key 以 handler 设置的值为准（覆盖语义）。

```
merged = request.headers ⊕ ctx.res.headers
```

**理由**：请求头回显是当前行为（用于 RPC ID 关联），不应破坏。覆盖语义最直觉，也最有用（middleware 可以预设、handler 可以覆盖）。

### 决策 5：DispatchFunc / Middleware 签名变更

```go
// 旧
type DispatchFunc func(conn *Connection, pkt packet.Packet) error
// 新
type DispatchFunc func(ctx HandlerContext) error
```

**理由**：如果 middleware 不使用 HandlerContext，则 `ctx.Set()` 传值机制无法工作，整个设计失去核心价值。签名统一后，middleware 和 handler 共享同一套上下文 API。

### 决策 6：内部通过类型断言分发

`Method` 和 `Notify` 的内部包装器分别将 `HandlerContext` 断言为 `*RequestContext` 或 `*NotifyContext`。这是安全的，因为路由器控制了 Context 的创建——OnRequestCB 只创建 RequestContext，OnNotifyCB 只创建 NotifyContext。

## 风险 / 权衡

- **[破坏性变更]** 所有 handler 和 middleware 签名需要更新 → 框架处于早期（仅 `sora-test` 示例），迁移范围可控。可通过编译错误引导逐个修改
- **[类型断言]** Middleware 中区分 Request/Notify 需要类型断言 `ctx.(*RequestContext)` → 文档和示例中明确推荐此模式。如果 middleware 只需要通用操作（读头、传值），无需断言
- **[每请求分配]** 每个请求创建新的 RequestContext/NotifyContext 对象（含 map 分配）→ 对于 RPC 框架可忽略不计，且 context 对象是短生命周期，GC 压力小
