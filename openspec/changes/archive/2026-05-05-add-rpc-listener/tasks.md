## 1. Transport 接口扩展

- [x] 1.1 在 `pkg/rpc/transport.go` 的 `Transport` interface 新增 `Handshake(ctx context.Context) (string, error)` 方法
- [x] 1.2 在 `pkg/rpc/transport/tcp/tcp_transport.go` 实现 `TCPTransport.Handshake`：从已建立的连接读取 codec 行（`"${codec}\n"`），写回确认行，返回 codec 名称

## 2. Connection server-side 初始化

- [x] 2.1 在 `pkg/rpc/connector.go` 实现 `Connection.Serve()`：设置 ctx，状态 Init → Connecting，调用 `transport.Handshake(ctx)` 获取 codec，通过 `GetCodec` 验证，设置 codec，启动 readLoop，状态 → Ready
- [x] 2.2 在 `pkg/rpc/connector.go` 实现 `Connection.SendResponse(ctx context.Context, res *packet.ResPacketData) error`：构造 `Packet{Opcode: PacketOpcodeResponse, Res: res}` 调用 `SendRaw`

## 3. TransportListener 接口与 TCP 实现

- [x] 3.1 在 `pkg/rpc/transport.go` 定义 `TransportListener` interface：`Accept(ctx context.Context) (*Connection, error)`、`Close() error`
- [x] 3.2 创建 `pkg/rpc/transport/tcp/tcp_listener.go`：定义 `TCPListener` struct（net.Listener + ConnectorOptions）
- [x] 3.3 实现 `NewTCPListener(addr string, opts TCPOptions, connOpts ConnectorOptions) (*TCPListener, error)`：调用 net.Listen 绑定地址
- [x] 3.4 实现 `TCPListener.Accept`：调用 net.Listener.Accept → 创建 TCPTransport（server mode，conn 已设置）→ NewConnection(transport, connOpts) → 返回 `*Connection`
- [x] 3.5 实现 `TCPListener.Close`：关闭底层 net.Listener
- [x] 3.6 添加 `var _ rpc.TransportListener = (*TCPListener)(nil)` 编译期接口检查

## 4. Listener 编排层

- [x] 4.1 重构 `pkg/rpc/listener.go`：定义 `ListenerState` 枚举（Init=1, Starting=2, Ready=3, Stopping=4, Stopped=5, Error=100）
- [x] 4.2 定义 `ListenerCallbacks` struct：`OnRequest`、`OnNotify`、`OnSessionOpen`、`OnSessionClose` 四个回调函数字段
- [x] 4.3 定义 `ListenerOptions` struct：嵌入 `ListenerInfo` 和 `ConnectorOptions`
- [x] 4.4 定义 `Listener` struct：TransportListener、options、callbacks、sessions map、LifeCycle、ctx/cancel
- [x] 4.5 实现 `NewListener(tl TransportListener, options ListenerOptions, callbacks ListenerCallbacks) *Listener`
- [x] 4.6 实现 `Listener.Start(ctx context.Context) error`：状态 Init → Starting，启动 acceptLoop goroutine，状态 → Ready，阻塞返回
- [x] 4.7 实现 acceptLoop：循环调用 `tl.Accept(ctx)`，对每个 Connection 调用 `Serve()`，成功后调用 `newConnector` 注册并触发 `OnSessionOpen`，失败触发 `OnSessionClose`
- [x] 4.8 实现 `newConnector(sessionId string, conn *Connection)`：注册到 sessions map，监听 Connection 生命周期（Error/Stopped 时清理 session 并触发 OnSessionClose）
- [x] 4.9 实现 `Listener.Stop()`：状态 → Stopping，cancel context，关闭 TransportListener，对所有 session 调用 Disconnect，状态 → Stopped
- [x] 4.10 实现 `Listener.CloseSession(sessionId string) error`：查找 session，调用 Disconnect，从 map 移除
- [x] 4.11 实现 `Listener.GetSession(sessionId string) (*Connection, bool)`

## 5. Connection readLoop 回调集成

- [x] 5.1 为 Connection 添加可选的 packet handler callback 字段（`OnRequest`、`OnNotify`），用于 Listener 路由
- [x] 5.2 修改 `handlePacket`：收到 Request packet 时调用 `OnRequest` callback（如果设置），收到 Notify packet 时调用 `OnNotify` callback（如果设置）

## 6. 集成验证

- [x] 6.1 更新 `cmd/sora-test/main.go` 或创建新测试文件：启动 Listener + Connector 端到端连接验证
- [x] 6.2 验证编译通过（`go build ./...`）
