## Why

sora-go-framework 的框架层（Runtime、Worker、Service、RPC、Provider 等）在运行时完全没有日志输出。基础设施 `pkg/logger` 已实现完善但无人调用。这导致运行时问题无法诊断：进程启动/关闭顺序不可见、服务生命周期不可追踪、RPC 连接错误无记录、goroutine panic 静默丢失。参考 sora-node/packages/framework（TypeScript 版）的做法，在框架各层添加结构化日志输出。

## What Changes

- 在 `pkg/runtime/runtime.go` 的 Startup/Shutdown 流程中添加生命周期日志（config-loaded、runtime-started、service/worker install/uninstall、discovery 连接/断开）
- 在 `pkg/runtime/runtime.go` 的 Startup 中添加 OS 信号处理（SIGINT/SIGTERM），记录日志后触发 Shutdown
- 在 `pkg/runner/woker.go` 的 Start/Stop 中添加 worker 生命周期日志
- 在 `pkg/runner/service.go` 的 Start/Stop/InstallListener 中添加 service 和 listener 生命周期日志
- 在 `pkg/rpc/connector.go` 中添加连接建立/断开/错误日志
- 在 `pkg/rpc/listener.go` 中添加 listener 启停和会话管理日志
- 在 `pkg/rpc/provider/provider.go` 中添加 sender 创建/移除日志
- 在 `pkg/rpc/provider/rpc_sender.go` 中添加连接循环日志
- 在 `pkg/discovery/` 的注册/注销操作中添加日志
- 在关键 goroutine 入口添加 `defer recover()` + error 日志，防止 panic 静默丢失

## Capabilities

### New Capabilities

- `framework-logging`: 框架层结构化日志输出，覆盖 Runtime、Worker/Service、RPC（Connector/Listener/Provider）、Discovery 全链路的生命周期和错误事件

### Modified Capabilities

（无需求变更，所有日志均为新增调用点）

## Impact

- **代码**: `pkg/runtime/`、`pkg/runner/`、`pkg/rpc/`、`pkg/discovery/` 目录下的文件
- **依赖**: 仅依赖已有的 `pkg/logger` 和 `pkg/runtime`（通过 `runtime.RT.FrameLogger` 和 `runtime.RT.RpcLogger`）
- **API**: 无破坏性变更，所有日志调用为纯副作用
- **运行时行为**: 进程现在会输出结构化日志到 stdout（通过已配置的 ConsoleOutput）
