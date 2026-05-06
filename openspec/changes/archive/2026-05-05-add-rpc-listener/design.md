## Context

sora-go-framework 是一个纯 Go 手写的 RPC 框架，已完成客户端 Connector 架构。当前状态：

- `Transport` interface 支持 client-side `Connect`（dial + handshake + zlib + length-prefix）
- `Connection` (Connector) 管理完整的客户端生命周期（Init → Connecting → Ready → Stopping → Stopped）
- `TCPTransport` 是唯一的 Transport 实现
- `ListenerInfo` 仅是数据结构（Protocol, Endpoint, Codecs, Labels），无服务端逻辑
- 服务端完全缺失：没有 accept、session 管理、request 路由

TCP 握手协议已在客户端实现：客户端发送 `"${codec}\n"`，服务端确认 `"${codec}\n"`。服务端需要对应的 accept + handshake 逻辑。

## Goals / Non-Goals

**Goals:**
- Transport interface 支持双向初始化：client-side Connect / server-side Handshake
- TransportListener interface 抽象服务端连接接受，Accept 返回 `*Connection`
- Connection 新增 `Serve()` 方法：server-side 初始化（Handshake → Ready → readLoop）
- Connection 新增 `SendResponse()` 方法：server reply
- Listener 编排服务端流程：acceptLoop、session 注册、callback 路由
- TCPListener 实现 TransportListener：构造时绑定地址，Accept 创建 TCPTransport + Connection
- Listener.Start 阻塞直到 Ready
- Server-side Connection 的 Ping/Pong disabled（仅被动回 pong）

**Non-Goals:**
- 不集成 runner 包（Listener 独立使用）
- 不实现内置 service registry（只暴露 callback）
- 不实现 WebSocket/HTTP 等其他 Transport 实现
- 不实现 codec 协商逻辑（Listener 不管客户端如何选择 codec，Connector 通过全局 registry 自行验证）
- 不实现请求-响应关联机制（Request/Response 的关联由上层通过 Headers 自行处理）

## Decisions

### D1: Transport.Handshake 方法签名

```go
type Transport interface {
    Connect(ctx context.Context, endpoint string, codec string) (string, error)
    Handshake(ctx context.Context) (string, error)
    Send(ctx context.Context, data []byte) error
    Recv(ctx context.Context) ([]byte, error)
    Close() error
}
```

**选择**: 新增 `Handshake(ctx) → (string, error)` 方法。

**替代方案**: 将 handshake 逻辑放在 TransportListener 层，Transport interface 不变。被否决——Connector 需要自己管理 codec（通过 Handshake 返回值获取客户端选择的 codec，然后到全局 registry 查找验证）。

**理由**: `Connect` 和 `Handshake` 是 Transport 的两种初始化路径，对称且互斥。`Connect` = dial + handshake（主动方），`Handshake` = handshake only（被动方，连接已由 TransportListener 建立）。

### D2: TransportListener.Accept 返回 *Connection

```go
type TransportListener interface {
    Accept(ctx context.Context) (*Connection, error)
    Close() error
}
```

**选择**: Accept 直接返回 `*Connection`，TCPListener 内部直接调用 `NewConnection` 构造。

**替代方案 A**: Accept 返回 `(Transport, error)`，Listener 负责包装成 Connection。被否决——用户要求 TransportListener 生成 connector。

**替代方案 B**: 通过工厂函数 `connFactory func(transport Transport) *Connection` 构造。被否决——增加了不必要的间接层，TCPListener 直接持有 `ConnectorOptions` 并调用 `NewConnection(transport, connOpts)` 更简单直接。

**实现**: TCPListener 持有 `ConnectorOptions`，Accept 内部：accept TCP → 创建 TCPTransport → `NewConnection(transport, connOpts)` → 返回 `*Connection`。

**理由**: TCPListener 本身就在 `transport/tcp` 包内，直接导入 `pkg/rpc` 的 `NewConnection` 不造成循环依赖（`pkg/rpc` 不导入 `transport/tcp`）。比工厂函数更直观。

### D3: Connection.Serve() server-side 初始化

```go
func (c *Connection) Serve() error
```

**选择**: 新增 `Serve()` 方法，server-side 初始化路径。

**流程**:
1. `ctx = context.Background()`
2. `LifeCycle: Init → Connecting`
3. `codecName, err := transport.Handshake(ctx)` — 等待客户端选择 codec
4. `codec := GetCodec(codecName)` — 全局 registry 查找验证
5. `LifeCycle: Connecting → Ready`
6. `go readLoop()`

**状态复用**: 使用现有的 `ConnectorStateConnecting` 状态。语义上 server 端不是 "connecting" 而是 "negotiating"，但为了避免状态枚举分裂，复用是可接受的。

**替代方案**: 新增 `ConnectorStateAccepting = 3`。被否决——增加状态枚举的复杂度不值得，Connecting 的语义（"正在建立通信"）在 server 侧也可接受。

**理由**: `Start` 和 `Serve` 是 Connection 的两种初始化入口，二选一。`Start` 用于 client（发起连接），`Serve` 用于 server（等待客户端）。命名遵循 Go 惯例（`http.Serve`）。

### D4: Listener 编排层

```go
type Listener struct {
    tl            TransportListener
    options       ListenerOptions
    callbacks     ListenerCallbacks
    sessions      map[string]*Connection
    sessionMu     sync.RWMutex
    LifeCycle     *utility.LifeCycle[ListenerState]
    ctx           context.Context
    cancel        context.CancelFunc
}
```

**状态机**: `Init(1) → Starting(2) → Ready(3) → Stopping(4) → Stopped(5)`，任意状态可进入 `Error(100)`。

**Start 阻塞**: Start 阻塞直到 Ready。TransportListener 在构造时已完成地址绑定，Start 启动 acceptLoop goroutine 后立即进入 Ready 状态。

**acceptLoop 流程**:
1. `conn, err := tl.Accept(ctx)`
2. `sessionId := uuid.New().String()`
3. `conn.Serve()` — Connector 自行完成 codec 协商
4. `newConnector(sessionId, conn)` — 注册 session
5. `callbacks.OnSessionOpen(conn)`
6. 监听 conn 生命周期，OnSessionClose 时清理

**callback 路由**: readLoop 中收到的 Request/Notify 通过 callback 传递给上层，Listener 不做业务路由。

### D5: Server-side Ping/Pong

Server-side Connection 的 `ConnectorOptions.Ping.Enabled = false`。Pinger 不启动，但 `handleCommand` 中收到 ping 时仍然回 pong（该逻辑在 `handleCommand` 中，不依赖 Pinger 是否启用）。

### D6: Connection.SendResponse

```go
func (c *Connection) SendResponse(ctx context.Context, res *packet.ResPacketData) error
```

基于 `SendRaw` 实现的便利方法，构造 `Packet{Opcode: PacketOpcodeResponse, Res: res}`。

## Risks / Trade-offs

- **[BREAKING API]** Transport interface 新增 `Handshake` 方法 → TCPTransport 需实现。当前只有一个实现，影响可控
- **[TransportListener 循环依赖]** TCPListener 直接调用 `NewConnection` → 导入 `pkg/rpc`。`pkg/rpc` 不导入 `transport/tcp`，无循环依赖
- **[Serve 错误时机]** Serve 内 Handshake 失败时 Connection 进入 Error 状态，但 Listener 的 acceptLoop 需要处理这种错误（不注册 session，触发 OnSessionClose）
- **[阻塞 Start]** Listener.Start 阻塞意味着调用方需要自行决定是否在 goroutine 中调用。这是有意设计——确保 Start 返回后 Listener 确实已就绪
- **[全局 Codec Registry]** Server-side Connection 通过全局 registry 验证 codec，如果客户端请求未注册的 codec，Handshake 会在 server 端写入确认后 Serve 仍会失败。考虑在 Transport.Handshake 层直接返回错误（写入 error 行）或 Connection.Serve 在 GetCodec 失败时关闭连接
