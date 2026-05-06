## Why

`baseWorker` 当前使用 `startupCtx`/`startupCancel` 仅覆盖 Startup 调用期间的 context 管理，Startup 返回后即被取消和置空。这导致：(1) Worker 运行期间没有统一的 context 用于传播取消信号；(2) `Executor` 拥有独立的 `context.Background()` 派生 ctx，与 Worker 生命周期脱节；(3) `Runner.Shutdown()` 无法获取任何 context 信息。应改为与 `Listener` 一致的全生命周期 context 模式，同时将 Executor 的等待能力吸收进 baseWorker 以简化架构。

## What Changes

- **BREAKING**: `baseWorker.Start()` 签名改为 `Start(ctx context.Context) error`，接受外部 context 用于级联控制
- 移除 `startupCtx`/`startupCancel` 字段，替换为 `ctx`/`cancel` 全生命周期 context
- 移除 `Executor` 字段，将 `sync.WaitGroup` + `atomic.Bool` 吸收到 `baseWorker` 中
- `Go(fn)` 直接使用 baseWorker 的 lifecycle ctx 和 WaitGroup
- `Stop()` 执行顺序改为：`running=false` → `cancel()` → `wg.Wait()` → `Shutdown()` → `Stopped`
- `Runner` 接口不变（`Shutdown()` 不接受 context）

## Capabilities

### New Capabilities
- `worker-lifecycle`: Worker 全生命周期 context 管理，包括 Start/Stop/Go 的 context 传播、任务追踪与优雅退出

### Modified Capabilities

## Impact

- `pkg/runner/woker.go` — 主要改动文件
- `pkg/runner/service.go` — 可能需要微调
- `pkg/utility/executor.go` — 不再被 baseWorker 引用，评估是否保留
- 无外部调用方（框架内部使用）
- Runner 接口签名不变，现有实现者无需修改
