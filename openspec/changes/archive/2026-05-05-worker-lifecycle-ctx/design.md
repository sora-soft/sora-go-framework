## Context

`pkg/runner` 中的 `baseWorker` 是框架中 Worker 和 Service 的基础结构。当前 context 管理方式存在以下问题：

- `startupCtx`/`startupCancel` 仅在 `Start()` 调用期间存活，Startup 返回后立即 cancel 并置 nil
- `Executor` 拥有独立的 `context.Background()` 派生 ctx，与 Worker 生命周期完全脱节
- `Executor` 从未在 `Start()` 中被启动（`Executor.Start()` 未被调用），`Go()` 实际静默丢弃任务
- `Runner.Shutdown()` 无法获取任何 context 信息
- 与 `pkg/rpc/listener.go` 中已验证的全生命周期 context 模式不一致

## Goals / Non-Goals

**Goals:**
- 统一 Worker 的 context 生命周期：从 `Start(ctx)` 创建到 `Stop()` 取消
- 吸收 Executor 的 WaitGroup + 守卫能力到 baseWorker，消除独立组件
- 所有 `Go()` 任务共享 lifecycle ctx，支持级联取消和优雅退出
- 与 Listener 模式保持一致的架构风格

**Non-Goals:**
- 修改 `Runner` 接口（`Shutdown()` 不加 context 参数）
- 修改 `Executor` utility（保留供其他使用者，仅 baseWorker 不再引用）
- 引入新的外部依赖
- 改变 Worker 的状态机（WorkerState 枚举不变）

## Decisions

### Decision 1: Start 接受外部 context

`Start()` 签名改为 `Start(ctx context.Context) error`。

**理由**: 与 Listener 模式一致；支持调用方通过父 context 控制生命周期；符合 Go context 传播惯例。lifecycle ctx 通过 `context.WithCancel(ctx)` 从父 ctx 派生。

### Decision 2: 吸收 Executor 到 baseWorker

移除 `Executor` 字段，直接在 baseWorker 中持有 `sync.WaitGroup` 和 `atomic.Bool`。

**理由**: Executor 的 ctx 独立于 lifecycle ctx 是设计缺陷；吸收后 ctx 统一；减少间接层；`Go()` 使用 lifecycle ctx 是唯一合理选择。

**替代方案**: 保留 Executor 但在 `Start()` 中从 lifecycle ctx 重建——增加复杂度但无额外收益。

### Decision 3: Stop 执行顺序

```
running = false
cancel()
wg.Wait()
LifeCycle → Stopping
Runner.Shutdown()
LifeCycle → Stopped
```

**理由**: cancel 先发出退出信号，wg.Wait 确保任务优雅完成，然后再做状态通知和最终清理。cancel 在 wg.Wait 之前确保任务能通过 ctx.Done() 感知退出信号。

**替代方案**: 先 SetState(Stopping) 再 cancel（Listener 的做法）——但 Listener 没有需要等待的内部任务，Worker 的 wg.Wait 需要取消信号先发出才能让任务退出。

### Decision 4: Shutdown 不接受 context

**理由**: 到 `Shutdown()` 执行时，所有任务已通过 cancel + Wait 退出，Shutdown 仅做同步资源清理。如需超时控制由调用方在 Stop() 外层包装。

## Risks / Trade-offs

- **[BREAKING] Start 签名变更** → 当前无外部调用方，影响为零。如未来有外部使用需文档说明。
- **Go() 在 Start 前调用会静默丢弃** → 与原 Executor 行为一致，acceptable。
- **cancel() 在 wg.Wait() 之前** → 要求 Go() 任务必须尊重 ctx.Done()，否则会死等。这是合理的约定。
- **Executor utility 闲置** → 保留在代码库中不删除，未来其他模块可能使用。
