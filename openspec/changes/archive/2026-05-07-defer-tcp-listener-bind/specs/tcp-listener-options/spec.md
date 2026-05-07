## MODIFIED Requirements

### Requirement: TCPListenerOptions 定义端口选择策略
系统 SHALL 提供 `TCPListenerOptions` struct，包含 `Host`（string）、`Port`（int）、`PortRange`（[]int）三个字段。`Port` 和 `PortRange` SHALL 互斥——仅可设置其中一个。

#### Scenario: 使用固定端口
- **WHEN** `TCPListenerOptions` 设置了 `Host` 和 `Port`，未设置 `PortRange`，调用 `StartListen(ctx)`
- **THEN** `TCPListener` SHALL 绑定到 `Host:Port`

#### Scenario: 使用端口范围
- **WHEN** `TCPListenerOptions` 设置了 `Host` 和 `PortRange`（长度为 2 的 `[min, max]`），未设置 `Port`，调用 `StartListen(ctx)`
- **THEN** `TCPListener` SHALL 从 `min` 端口开始尝试绑定，每次失败后步进 `rand(1, 5)`，直到成功或超出 `max`

#### Scenario: 同时设置 Port 和 PortRange
- **WHEN** `TCPListenerOptions` 同时设置了 `Port` 和 `PortRange`
- **THEN** `NewTCPListener` SHALL 返回错误

#### Scenario: Port 和 PortRange 都未设置
- **WHEN** `TCPListenerOptions` 的 `Port` 为零值且 `PortRange` 为空
- **THEN** `NewTCPListener` SHALL 返回错误

### Requirement: 端口范围扫描行为
当使用 `PortRange` 模式时，`TCPListener.StartListen` SHALL 从 `PortRange[0]` 开始，以 `rand(1, 5)` 为步进尝试绑定端口。步进后端口超过 `PortRange[1]` 时 SHALL 返回错误。

#### Scenario: 首次尝试即成功
- **WHEN** `PortRange` 为 `[10000, 10100]`，端口 10000 可用，调用 `StartListen(ctx)`
- **THEN** `TCPListener` SHALL 成功绑定到 10000

#### Scenario: 需要多次尝试后成功
- **WHEN** `PortRange` 为 `[10000, 10100]`，端口 10000 和 10003 已被占用，10007 可用，调用 `StartListen(ctx)`
- **THEN** `TCPListener` SHALL 成功绑定到 10007（或范围内其他可用端口）

#### Scenario: 整个范围均不可用
- **WHEN** `PortRange` 为 `[10000, 10005]`，所有端口均被占用，调用 `StartListen(ctx)`
- **THEN** `StartListen` SHALL 返回错误

### Requirement: 移除 TCPOptions
系统 SHALL 移除 `TCPOptions` struct。原 `TCPOptions` 中的传输参数（`MaxRetries=3`、`InitialDelay=500ms`、`MaxDelay=8s`、`ConnectTimeout=5s`）SHALL 硬编码进 `TCPTransport`。

#### Scenario: TCPTransport 使用硬编码默认值
- **WHEN** 创建 `TCPTransport` 或 `ServerTCPTransport`
- **THEN** 重连/超时参数 SHALL 使用硬编码默认值，不接受外部配置

### Requirement: NewTCPListener 签名变更
`NewTCPListener` SHALL 接受 `TCPListenerOptions` 参数，移除原有的 `addr string` 和 `TCPOptions` 参数。`NewTCPListener` SHALL 不执行端口绑定，仅校验参数并保存配置。

#### Scenario: 使用新签名创建 listener（无副作用）
- **WHEN** 调用 `NewTCPListener(TCPListenerOptions{Host: "127.0.0.1", Port: 8080})`
- **THEN** SHALL 成功创建 `*TCPListener` 实例，但不绑定端口

#### Scenario: 端口绑定延迟到 StartListen
- **WHEN** 调用 `NewTCPListener(TCPListenerOptions{Host: "127.0.0.1", Port: 8080})` 后调用 `StartListen(ctx)`
- **THEN** SHALL 在 `StartListen` 调用时绑定到 `127.0.0.1:8080`
