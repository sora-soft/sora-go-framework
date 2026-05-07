## 为什么

Provider 的 Start 方法可以被多次调用（被多个调用者引用），但当前实现没有引用计数保护。每次调用 Start 都会重新创建 context/cancel 并启动新的 watchLoop goroutine，导致 context 泄漏、goroutine 泄漏和并发竞态。component 包已有成熟的引用计数模式（baseComponent），Provider 应对齐这一模式。

## 变更内容

- 为 Provider 添加引用计数器（refCount），使 Start/Stop 支持多次调用配对
- **BREAKING**: `Provider.Start` 签名从 `Start(ctx context.Context)` 变更为 `Start(ctx context.Context) error`
- **BREAKING**: `Provider.Stop` 签名从 `Stop()` 变更为 `Stop() error`
- Start 在 refCount > 0 时仅递增计数，忽略新传入的 ctx
- Stop 仅在 refCount 降至 0 时执行实际清理

## 功能

### 新增功能

- `provider-lifecycle`: Provider 的引用计数生命周期管理，包括 Start/Stop 的幂等性、配对调用和资源保护

### 修改功能

_(无现有功能规范需要修改)_

## 影响

- `pkg/rpc/provider/provider.go` — 核心变更，添加 refCount 字段，修改 Start/Stop 实现
- `pkg/rpc/provider/rpc_sender.go` — 无需变更，RpcSender 的生命周期由 Provider 管理
- 所有调用 `Provider.Start` 和 `Provider.Stop` 的代码需适配新签名
