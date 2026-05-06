## Why

当前 RPC 框架只有客户端 Connector 实现，缺少服务端 Listener。TCP 握手协议已在 Connector 和 TCPTransport 中定义（客户端发送 codec、服务端确认），但服务端没有对应的 accept 和 session 管理。需要实现 Listener 以支持完整的 RPC 双向通信。

## What Changes

- **BREAKING**: `Transport` interface 新增 `Handshake(ctx) (string, error)` 方法，用于 server-side codec 协商
- 新增 `TransportListener` interface：`Accept(ctx) → (*Connection, error)`，负责接受原始连接并构造 Connection
- 新增 `Connection.Serve()` 方法：server-side 初始化路径，调用 `Transport.Handshake` 完成 codec 协商后进入 Ready + readLoop
- 新增 `Connection.SendResponse(ctx, *ResPacketData)` 方法：server reply 便利方法
- 新增 `Listener` struct：服务端编排层，管理 acceptLoop、session 注册、callback 路由
- 新增 `TCPListener`：`TransportListener` 的 TCP 实现，构造时绑定地址，Accept 创建 TCPTransport + Connection
- `Listener.Start(ctx)` 阻塞直到 Ready，`Stop()` 关闭所有 session
- Listener 通过 callback 暴露 `OnRequest`、`OnNotify`、`OnSessionOpen`、`OnSessionClose`
- Server-side Connection 的 Ping/Pong：disabled（不主动 ping），仅被动回 pong

## Capabilities

### New Capabilities
- `listener-lifecycle`: Listener 状态机（Init → Starting → Ready → Stopping → Stopped），阻塞式 Start，Stop 关闭所有 session
- `session-management`: acceptLoop 接受连接、构造 Connection、通过 newConnector(sessionId, conn) 注册 session，session 生命周期回调
- `transport-listener`: TransportListener interface 定义，Accept 返回 *Connection，TCP 实现包含地址绑定和连接接受

### Modified Capabilities
- `transport-interface`: 新增 `Handshake(ctx) (string, error)` 方法，Transport 支持双向初始化（Connect client-side / Handshake server-side）
- `connector-lifecycle`: 新增 `Serve()` server-side 初始化路径和 `SendResponse` 方法

## Impact

- **BREAKING API**: `Transport` interface 新增方法，所有 Transport 实现（当前仅 TCPTransport）需实现 `Handshake`
- 新增文件：`pkg/rpc/transport/tcp/tcp_listener.go`（TCPListener 实现）
- 修改文件：`pkg/rpc/transport.go`（Transport interface 加 Handshake）、`pkg/rpc/connector.go`（加 Serve/SendResponse）、`pkg/rpc/transport/tcp/tcp_transport.go`（加 Handshake 实现）、`pkg/rpc/listener.go`（Listener struct）
- `cmd/sora-test/main.go` 可用于端到端验证
- 无外部依赖变更
