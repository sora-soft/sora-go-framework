## 1. TransportListener interface 变更

- [x] 1.1 在 `TransportListener` interface 中新增 `StartListen(ctx context.Context) error` 方法
- [x] 1.2 验证编译通过（此时 TCPListener 会报错，预期行为）

## 2. TCPListener 重构

- [x] 2.1 `NewTCPListener` 移除端口绑定逻辑（`net.Listen`），仅校验参数、保存 opts、返回实例
- [x] 2.2 `TCPListener` 新增 `listener net.Listener` 字段和 `started bool` 标志（或 `sync.Once`）
- [x] 2.3 实现 `StartListen(ctx context.Context) error`，将原 `NewTCPListener` 中的端口绑定逻辑移入
- [x] 2.4 `StartListen` 支持幂等——已绑定时返回 nil
- [x] 2.5 `Accept` 在 `listener` 为 nil（未 StartListen）时返回明确错误
- [x] 2.6 `Close` 处理 `listener` 为 nil 的情况（未绑定时 Close 不报错）

## 3. Listener.Start 集成

- [x] 3.1 `Listener.Start()` 在 `SetState(Starting)` 之后、`acceptLoop` 之前调用 `tl.StartListen(ctx)`
- [x] 3.2 `StartListen` 失败时设置 Error 状态并返回错误，不启动 acceptLoop

## 4. 调用方适配

- [x] 4.1 更新 `cmd/test/main.go`，移除 `NewTCPListener` 的端口绑定依赖（构造后无需检查端口是否可用）
- [x] 4.2 确认所有 `TransportListener` 实现者已实现 `StartListen`

## 5. 验证

- [x] 5.1 编译通过
- [x] 5.2 运行现有测试
- [x] 5.3 `cmd/test/main.go` 端到端验证
