## ADDED Requirements

### Requirement: TransportListener interface
系统 SHALL 提供 `TransportListener` interface，包含 `Accept` 和 `Close` 两个方法。TransportListener 负责绑定地址、接受原始连接、构造 `*Connection` 返回。

#### Scenario: TransportListener interface 签名
- **WHEN** 定义 TransportListener interface
- **THEN** 包含 `Accept(ctx context.Context) (*Connection, error)`、`Close() error`

### Requirement: Accept 返回 *Connection
TransportListener.Accept SHALL 将底层 transport 包装为 `*Connection` 并返回。TCPListener 通过持有 `ConnectorOptions` 直接调用 `NewConnection` 构造。

#### Scenario: Accept 创建 Connection
- **WHEN** 调用 Accept(ctx) 且有新的底层连接
- **THEN** 创建 Transport（如 TCPTransport），调用 NewConnection(transport, connOpts)，返回 *Connection

#### Scenario: Accept context 取消
- **WHEN** 调用 Accept(ctx) 且 ctx 被取消
- **THEN** 返回 context 错误

### Requirement: TCPListener 实现
系统 SHALL 提供 `TCPListener` 作为 TransportListener 的 TCP 实现。TCPListener 在构造时绑定 TCP 地址。

#### Scenario: TCPListener 构造
- **WHEN** 调用 `NewTCPListener(addr, tcpOpts, connOpts)`
- **THEN** 绑定 TCP 地址，返回 *TCPListener 实例

#### Scenario: TCPListener.Accept 流程
- **WHEN** 调用 TCPListener.Accept(ctx)
- **THEN** 调用 net.Listener.Accept() 获取 net.Conn，创建 TCPTransport（conn 已设置），调用 NewConnection(transport, connOpts) 返回 *Connection

#### Scenario: TCPListener.Close
- **WHEN** 调用 TCPListener.Close()
- **THEN** 关闭底层 net.Listener
