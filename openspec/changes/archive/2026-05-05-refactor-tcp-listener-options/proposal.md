## Why

多实例部署时多个进程共享同一份配置文件，当前 `TCPListener` 只支持绑定固定地址（`addr string`），导致端口冲突。需要支持端口范围扫描，让每个实例自动选择可用端口。

## What Changes

- **BREAKING**: 移除 `TCPOptions` struct，将其字段硬编码进 `TCPTransport`
- **BREAKING**: `NewTCPListener` 签名变更：`addr string` 参数替换为 `TCPListenerOptions`
- 新增 `TCPListenerOptions` struct，包含 `Host`、`Port`（固定端口）、`PortRange`（端口范围扫描，二选一）
- `TCPListener` 在 `PortRange` 模式下从最小端口开始，以 `rand(1, 5)` 步进尝试绑定，直到成功或超出范围返回错误

## Capabilities

### New Capabilities
- `tcp-listener-options`: TCPListener 的选项配置，包含端口选择策略（固定端口 / 端口范围扫描）

### Modified Capabilities
<!-- 无现有 spec 需要修改 -->

## Impact

- `pkg/rpc/transport/tcp/tcp_listener.go`: 构造函数签名变更，新增端口扫描逻辑
- `pkg/rpc/transport/tcp/tcp_transport.go`: 移除 `TCPOptions`，硬编码传输层参数
- `cmd/sora-test/main.go`: 调用方需适配新签名
