## Why

`NewTCPListener` 在构造时立即绑定 TCP 端口（调用 `net.Listen`），这是一个有副作用的构造函数。这导致：
1. 构造与资源分配耦合 — 无法先创建对象再决定何时绑定
2. 测试困难 — 构造即占用端口，无法在单元测试中构造但不绑定
3. 与 `Listener.Start()` 的职责重叠 — `Start` 应负责将 listener 从"已创建"推进到"已就绪"，但端口绑定发生在 Start 之前

## What Changes

- **BREAKING** `TransportListener` interface 新增 `StartListen(ctx context.Context) error` 方法，将地址绑定从构造函数延迟到 `StartListen`
- `NewTCPListener` 不再绑定端口，仅保存配置，返回未绑定的 `*TCPListener`
- `Listener.Start()` 在启动 `acceptLoop` 之前先调用 `tl.StartListen(ctx)`
- `TCPListener.Accept()` 在未调用 `StartListen` 时返回错误
- 更新 `cmd/test/main.go` 等调用方代码

## Capabilities

### New Capabilities

_(无新增能力)_

### Modified Capabilities

- `transport-listener`: `TransportListener` interface 新增 `StartListen` 方法；`TCPListener` 构造不再绑定端口，绑定延迟到 `StartListen`
- `listener-lifecycle`: `Listener.Start()` 需在 `acceptLoop` 之前调用 `tl.StartListen(ctx)`；启动失败时状态回退到 Error
- `tcp-listener-options`: `NewTCPListener` 签名变更，不再执行端口绑定，端口绑定逻辑移入 `StartListen`

## Impact

- `pkg/rpc/transport.go` — `TransportListener` interface 新增方法
- `pkg/rpc/transport/tcp/tcp_listener.go` — `TCPListener` 重构
- `pkg/rpc/listener.go` — `Start()` 方法增加 `tl.StartListen` 调用
- `cmd/test/main.go` — 调用方适配
- 所有 `TransportListener` 实现者需新增 `StartListen` 方法（目前仅 TCP）
