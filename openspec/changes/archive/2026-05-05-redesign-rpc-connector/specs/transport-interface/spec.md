## ADDED Requirements

### Requirement: Transport interface defines pluggable wire protocol
系统 SHALL 提供 `Transport` interface，包含 `Connect`、`Send`、`Recv`、`Close` 四个方法。Transport 实现完全拥有 wire protocol 的所有细节（连接建立、handshake、封包/解包）。

#### Scenario: Transport interface signature
- **WHEN** 定义 Transport interface
- **THEN** 包含 `Connect(ctx context.Context, endpoint string, codec string) (string, error)`、`Send(ctx context.Context, data []byte) error`、`Recv(ctx context.Context) ([]byte, error)`、`Close() error`

### Requirement: Connect 返回服务方确认的 codec
`Connect` 方法 SHALL 返回 `(confirmedCodec string, err error)`。`confirmedCodec` 是服务方在 handshake 中回复的 codec 名称。

#### Scenario: handshake 成功且 codec 匹配
- **WHEN** 调用 `Connect(ctx, "host:port", "json")` 且服务方回复 `"json\n"`
- **THEN** 返回 `("json", nil)`

#### Scenario: handshake 成功但 codec 不匹配
- **WHEN** 调用 `Connect(ctx, "host:port", "json")` 且服务方回复 `"msgpack\n"`
- **THEN** 返回 `("", error)`

#### Scenario: 连接失败
- **WHEN** 调用 `Connect` 且底层网络不可达
- **THEN** 返回 `("", error)`

### Requirement: 所有 I/O 操作支持 context 传播
`Connect`、`Send`、`Recv` SHALL 接受 `context.Context` 参数。当 context 被取消时，操作 SHALL 立即终止并返回 `context.Cause(ctx)` 或 `ctx.Err()`。

#### Scenario: Connect 期间 context 取消
- **WHEN** 调用 `Connect(ctx, ...)` 且 ctx 被取消
- **THEN** 立即终止连接尝试并返回 context 错误

#### Scenario: Recv 期间 context 取消
- **WHEN** 调用 `Recv(ctx)` 阻塞等待数据时 ctx 被取消
- **THEN** 返回 context 错误

### Requirement: TCPTransport 默认实现
系统 SHALL 提供 `TCPTransport` 作为默认 Transport 实现，包含 TCP 连接、handshake 协议、zlib 压缩和 length-prefix 封包。

#### Scenario: TCP handshake 协议（发起方）
- **WHEN** TCPTransport 执行 Connect
- **THEN** 向服务方发送 `"${codec}\n"`，等待服务方回复 `"${codec}\n"`，验证匹配

#### Scenario: TCP 封包协议
- **WHEN** TCPTransport 执行 Send
- **THEN** 将数据 zlib 压缩后以 `[4字节 big-endian uint32 长度][压缩数据]` 格式发送

#### Scenario: TCP 解包协议
- **WHEN** TCPTransport 执行 Recv
- **THEN** 读取 4 字节长度头，读取对应长度的压缩数据，zlib 解压后返回

### Requirement: TCPTransport 可配置指数退避重试
TCPTransport SHALL 在 Connect 失败时自动重试。重试策略通过 `TCPOptions` 配置，包含 `MaxRetries`、`InitialDelay`、`MaxDelay`、`ConnectTimeout`。延迟公式 SHALL 为 `min(InitialDelay * 2^attempt, MaxDelay)`。

#### Scenario: 首次连接失败后自动重试
- **WHEN** TCPTransport 首次 dial 失败且 MaxRetries >= 2
- **THEN** 等待 `InitialDelay` 后重试

#### Scenario: 重试次数耗尽
- **WHEN** TCPTransport 重试次数达到 MaxRetries 仍失败
- **THEN** 返回最后一次的错误

#### Scenario: 重试期间 context 取消
- **WHEN** TCPTransport 在重试等待期间收到 ctx 取消信号
- **THEN** 立即停止重试并返回 context 错误

#### Scenario: 指数退避延迟
- **WHEN** MaxRetries=5, InitialDelay=500ms, MaxDelay=8s
- **THEN** 重试延迟依次为 500ms, 1s, 2s, 4s, 8s
