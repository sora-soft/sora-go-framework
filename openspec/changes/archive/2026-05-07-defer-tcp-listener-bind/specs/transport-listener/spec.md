## MODIFIED Requirements

### Requirement: TransportListener interface
系统 SHALL 提供 `TransportListener` interface，包含 `StartListen`、`Accept` 和 `Close` 三个方法。TransportListener 负责延迟绑定地址、接受原始连接、构造 `*Connection` 返回。

#### Scenario: TransportListener interface 签名
- **WHEN** 定义 TransportListener interface
- **THEN** 包含 `StartListen(ctx context.Context) error`、`Accept(ctx context.Context) (*Connection, error)`、`Close() error`

### Requirement: StartListen 延迟绑定地址
`StartListen(ctx context.Context) error` SHALL 执行底层地址绑定。调用前 `Accept` SHALL 返回错误。`StartListen` SHALL 支持幂等调用——多次调用时，若已成功绑定，SHALL 返回 nil。

#### Scenario: StartListen 绑定成功
- **WHEN** 调用 `StartListen(ctx)` 且绑定成功
- **THEN** 返回 nil，后续 `Accept` 调用正常工作

#### Scenario: StartListen 绑定失败
- **WHEN** 调用 `StartListen(ctx)` 且绑定失败（如端口不可用）
- **THEN** 返回错误，`Accept` 继续返回错误

#### Scenario: StartListen 幂等
- **WHEN** 调用 `StartListen(ctx)` 且已成功绑定过
- **THEN** 返回 nil，不重复绑定

#### Scenario: Accept 在 StartListen 之前调用
- **WHEN** 调用 `Accept(ctx)` 且 `StartListen` 未被调用
- **THEN** 返回错误

### Requirement: Accept 返回 *Connection
TransportListener.Accept SHALL 将底层 transport 包装为 `*Connection` 并返回。TCPListener 通过持有 `ConnectorOptions` 直接调用 `NewConnection` 构造。

#### Scenario: Accept 创建 Connection
- **WHEN** 调用 Accept(ctx) 且有新的底层连接
- **THEN** 创建 Transport（如 TCPTransport），调用 NewConnection(transport, connOpts)，返回 *Connection

#### Scenario: Accept context 取消
- **WHEN** 调用 Accept(ctx) 且 ctx 被取消
- **THEN** 返回 context 错误

### Requirement: TCPListener 实现
系统 SHALL 提供 `TCPListener` 作为 TransportListener 的 TCP 实现。TCPListener 在 `NewTCPListener` 构造时仅保存配置，不执行端口绑定。端口绑定 SHALL 延迟到 `StartListen` 调用时执行。

#### Scenario: TCPListener 构造（无副作用）
- **WHEN** 调用 `NewTCPListener(opts)`
- **THEN** 仅校验参数并保存配置，不调用 `net.Listen`，返回 `*TCPListener` 实例

#### Scenario: TCPListener.StartListen 绑定端口
- **WHEN** 调用 `TCPListener.StartListen(ctx)`
- **THEN** 执行 `net.Listen("tcp", addr)` 绑定 TCP 地址

#### Scenario: TCPListener.Accept 流程
- **WHEN** 调用 TCPListener.Accept(ctx) 且已 StartListen
- **THEN** 调用 net.Listener.Accept() 获取 net.Conn，创建 TCPTransport，调用 NewConnection(transport, connOpts) 返回 *Connection

#### Scenario: TCPListener.Close
- **WHEN** 调用 TCPListener.Close()
- **THEN** 关闭底层 net.Listener（若已绑定）
