## Context

sora-go-framework 是 sora-node/packages/framework 的 Go 语言重写版本。两个项目在架构上对齐：Runtime 管理 Node/Service/Worker 生命周期，RPC 层通过 Connector/Listener/Provider 实现 P2P 通信，Discovery 负责服务注册与发现。

Go 版本的 `pkg/logger` 已完整实现（Logger、LoggerData、ConsoleOutput、LogLevel），且 `runtime.RT` 已持有 `FrameLogger` 和 `RpcLogger` 实例。但框架各层代码从未调用它们。

TS 版本的日志模式：所有框架层通过 `Runtime.frameLogger` / `Runtime.rpcLogger` 全局访问，统一使用 `category` + `{event, ...context}` 的结构化 content 格式，输出为 `timeString,level,identify,category,position,content`。

## Goals / Non-Goals

**Goals:**
- 在框架所有关键路径添加结构化日志，与 TS 版本输出信息对齐
- 覆盖：启动/关闭、服务/Worker 生命周期、RPC 连接管理、Discovery 注册、错误与 panic
- 日志内容包含 event 标识和上下文数据（name、id、error 等），便于排查
- 添加 goroutine panic 恢复，防止静默崩溃

**Non-Goals:**
- 不修改 logger 基础设施（Logger、ConsoleOutput、LoggerData 等已完善）
- 不添加文件日志、远程日志等新 Output
- 不改变现有 API 签名或行为
- 不添加 trace/distributed tracing（超出范围）

## Decisions

### D1: 通过 runtime.RT.FrameLogger / RpcLogger 全局访问

**决策**: 所有框架层日志统一通过 `runtime.RT.FrameLogger` 和 `runtime.RT.RpcLogger` 访问。

**替代方案**: 在 BaseWorker/BaseService 结构体中持有 logger 引用。 rejected — 增加每个实例的内存开销，且与 TS 版本的全局 Logger 模式不一致。

**理由**: Go 项目的 `runtime.RT` 是全局单例，FrameLogger 和 RpcLogger 已在初始化时创建。与 TS 版本 `Runtime.frameLogger` / `Runtime.rpcLogger` 静态属性对齐。

### D2: Category 命名与 TS 对齐

**决策**: category 字符串直接沿用 TS 版本：
- `runtime` — Runtime 生命周期（startup、shutdown、signal、config）
- `connector` — RPC Connector 连接事件
- `listener` — RPC Listener 会话事件
- `provider.<name>` — Provider sender 管理事件
- `discovery` — Discovery 注册/注销事件

**理由**: 与 TS 版本日志输出格式一致，方便跨语言项目统一监控。

### D3: 信号处理在 runtime.Startup() 中注册

**决策**: 在 `runtime.Startup()` 中注册 `SIGINT`/`SIGTERM` handler，记录日志后调用 `runtime.Shutdown()`。

**替代方案**: 留给使用者在 main 中处理。rejected — 框架应管理自身生命周期，TS 版本也在 Runtime 中处理信号。

### D4: Goroutine panic 恢复 + error 日志

**决策**: 在框架启动的每个 goroutine 入口添加 `defer recoverPanic()` 模式：
```go
defer func() {
    if r := recover(); r != nil {
        runtime.RT.FrameLogger.Error("runtime", fmt.Errorf("%v", r), map[string]any{
            "event": "goroutine-panic",
            "recover": r,
        })
    }
}()
```

**替代方案**: 不恢复，让 panic 崩溃进程。rejected — TS 版本捕获 uncaughtException/unhandledRejection 并记录日志，Go 版应提供等效保护。

### D5: 日志级别使用规则

**决策**: 级别选择规则：
- `Debug` — 不使用（框架日志至少 Info 级别）
- `Info` — 流程开始/进行中（service-starting、worker-starting）
- `Success` — 流程完成（service-started、runtime-started、listener-started）
- `Warn` — 非致命异常（connector-response-not-enabled、parse-body-failed）
- `Error` — 操作失败（install-service-start、handle-command-error）
- `Fatal` — 进程级致命错误（connect-discovery 失败、uncaught-exception）

**理由**: 与 TS 版本的 LogLevel 枚举和使用语义完全对齐。

## Risks / Trade-offs

**[性能] 日志在高频路径上的开销** → ConsoleOutput 已使用 JSON 序列化，日志仅在关键生命周期节点触发，不在热路径（如每个 RPC packet）上记录。风险低。

**[全局单例] runtime.RT 初始化时序** → 如果某模块在 `runtime.RT` 初始化前尝试记录日志会 panic。但当前所有模块在 Runtime.Startup 之后才运行，风险可控。

**[信号处理] 覆盖使用者的信号处理** → 如果使用者也需要处理 SIGINT/SIGTERM，可能与 Runtime 的 handler 冲突。Go 的 `signal.Notify` 支持多个 handler，但 Shutdown 只应执行一次。通过 `sync.Once` 保护。
