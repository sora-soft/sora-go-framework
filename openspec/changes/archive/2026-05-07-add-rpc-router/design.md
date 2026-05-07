## Context

当前 RPC 层的 `ListenerCallbacks` 和 `Connection` 上的 `OnRequest` / `OnNotify` 是扁平回调，接收原始 `packet.Packet`。使用者需自行按 `pkt.Method` 分发、手动 `packet.Decode[T]()` 解码、手动构造并回传 Response。Packet 中已有 `Method` 和 `Service` 字段，但无路由机制。

约束：
- Listener 只收同一 service 的请求，路由只需 `Method` 一级
- Router 为纯包装层，不修改 `Listener` / `Connection` / `ListenerCallbacks` 任何现有代码
- Response payload 格式已确定：`{error: IPayloadError | null, result: T | null}`
- 现有 `errorx.Error` 结构与 `IPayloadError` 字段一一对应

## Goals / Non-Goals

**Goals:**
- 提供基于 `Method` 的一级路由，自动按 `pkt.Method` 分发到强类型 handler
- `Method[Req, Resp]` 注册 Request handler，`Notify[Msg]` 注册 Notify handler
- Router 同时输出 `func(conn, pkt)` 给 Listener 侧 (`ListenerCallbacks`) 和 Connector 侧 (`Connection.OnRequest/OnNotify`)
- 自动 Response 封装：成功时 `{error: null, result: <encoded>}`，失败时 `{error: {...}, result: null}`
- 未注册 method 自动返回 `ERR_METHOD_NOT_FOUND` 错误 Response
- 中间件链 `router.Use(middleware)`，在路由分发前执行
- Codec 从入站 `Packet` 上获取，Router 不持有 codec

**Non-Goals:**
- 不做 Service + Method 两级路由
- 不修改现有 `Listener`、`Connection`、`ListenerCallbacks` 代码
- 不做客户端（Provider）侧的请求构造路由——Provider 的 `CallRpc` 已直接指定 method
- 不做流式（streaming）RPC
- 不引入外部依赖

## Decisions

### D1: 注册时闭包擦除泛型

Go 泛型无法将不同 `Req`/`Resp` 类型的 handler 存入同一 map。解法是在 `Method[Req, Resp]()` 注册时，闭包捕获泛型参数，将 handler 包装为 `func(conn *Connection, pkt packet.Packet) error` 存入 `methodTable`。

```go
func Method[Req any, Resp any](r *Router, method string, handler func(conn *Connection, req Req) (Resp, error)) {
    r.methodTable[method] = func(conn *Connection, pkt packet.Packet) error {
        req, err := packet.Decode[Req](pkt)
        if err != nil { return err }
        resp, err := handler(conn, req)
        // ... encode & send response
    }
}
```

**替代方案**: 用 `reflect` 或 `any` + 运行时类型断言 —— 性能差且类型不安全。闭包方案零反射、编译期类型安全。

### D2: Method 与 Notify 分表存储

```go
type Router struct {
    methodTable map[string]func(conn *Connection, pkt packet.Packet) error
    notifyTable map[string]func(conn *Connection, pkt packet.Packet)
    middlewares []Middleware
    logger      Logger
}
```

- `methodTable`: Request handler，返回 error，Router 自动发 Response
- `notifyTable`: Notify handler，无返回值，错误仅记日志

**替代方案**: 统一一张表 + `NoResponse` 哨兵类型 —— 增加不必要的复杂度，语义不如分表清晰。

### D3: 中间件在路由分发层面 wrap

```go
type DispatchFunc func(conn *Connection, pkt packet.Packet) error
type Middleware  func(next DispatchFunc) DispatchFunc
```

Middleware 操作原始 Packet（可检查 Method、Headers），路由和 decode 在 middleware 链最内层。Middleware 同时作用于 Request 和 Notify。

**替代方案**: 逐个 handler wrap —— Go 泛型难以实现统一的 handler wrapper，且无法在中间件中访问原始 Packet。

### D4: Response 封装格式

```go
type ResPayload[T any] struct {
    Error  *PayloadError `json:"error"`
    Result T             `json:"result"`
}

type PayloadError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Level   int    `json:"level"`
    Name    string `json:"name"`
    Args    any    `json:"args"`
}
```

- handler 返回 `*errorx.Error` → 直接映射字段
- handler 返回普通 `error` → 兜底为 `{code: "ERR_INTERNAL", level: UNEXPECTED, ...}`
- Codec 从 `pkt.codec` 取，Router 无 codec 依赖

### D5: 错误处理策略

| 场景 | 处理 |
|------|------|
| Method handler 返回 `*errorx.Error` | 映射为 PayloadError，发错误 Response |
| Method handler 返回普通 error | 兜底 ERR_INTERNAL，发错误 Response |
| Method 未注册 | ERR_METHOD_NOT_FOUND，发错误 Response |
| Decode 失败 | ERR_DECODE_FAILED，发错误 Response |
| Notify handler 返回 error | 仅记日志，不发 Response |
| Notify 未注册 | 仅记日志（Warn），不发 Response |

## Risks / Trade-offs

- **[风险] Router 出错导致连接断开** → Router 内部 recover panic，转为错误 Response（Request）或日志（Notify），不向上抛
- **[权衡] Middleware 无法感知强类型** → Middleware 工作在 Packet 层面，无法访问 decoded struct。这是 Go 泛型限制，可通过在 ctx/headers 中传递元数据缓解
- **[权衡] Response headers 来源** → Response 复用 Request Packet 的 Headers，保持请求-响应关联。如有需要 handler 可通过扩展 mechanism 修改
