## 1. 重构 TCPTransport：移除 TCPOptions

- [x] 1.1 在 `tcp_transport.go` 中移除 `TCPOptions` struct 和 `DefaultTCPOptions()` 函数
- [x] 1.2 将 `MaxRetries=3`、`InitialDelay=500ms`、`MaxDelay=8s`、`ConnectTimeout=5s` 硬编码为 `TCPTransport` 内部常量
- [x] 1.3 更新 `TCPTransport` struct 移除 `opts` 字段，所有方法改用内部常量
- [x] 1.4 更新 `NewTCPTransport` 和 `NewServerTCPTransport` 签名，移除 `opts TCPOptions` 参数

## 2. 新增 TCPListenerOptions 和端口扫描逻辑

- [x] 2.1 在 `tcp_listener.go` 中定义 `TCPListenerOptions` struct（`Host string`、`Port int`、`PortRange []int`）
- [x] 2.2 实现参数校验：`Port` 和 `PortRange` 二选一互斥，都不设置返回错误
- [x] 2.3 实现 `PortRange` 模式的端口扫描逻辑：从 min 开始，步进 `rand(1, 5)`，超出 max 返回错误
- [x] 2.4 更新 `NewTCPListener` 签名为 `(opts TCPListenerOptions, connOpts rpc.ConnectorOptions)`
- [x] 2.5 更新 `TCPListener` struct，将 `opts TCPOptions` 替换为 `listenerOpts TCPListenerOptions`

## 3. 更新调用方

- [x] 3.1 更新 `cmd/sora-test/main.go` 中的 `NewTCPListener` 和 `NewTCPTransport` 调用，适配新签名
