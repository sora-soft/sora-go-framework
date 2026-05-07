## Why

当前 `provider.Provider` 是一个具体 struct，无法替换或 mock。同时，Worker/Service 层缺少统一注册 Provider 的机制——用户需要自行管理 Provider 的启动和停止生命周期。引入 Provider interface 并在 runner 层提供 `RegisterProvider` 方法，可以让 Provider 的生命周期管理与 Component 保持一致的模式，支持 Provider 实例在多个 Worker/Service 间共享（通过引用计数），并使框架更加可测试和可扩展。

## What Changes

- 将 `provider.Provider` 从具体 struct 重构为 interface，包含 `Start`、`Stop`、`CallRpc` 方法
- 将具体实现重命名为 `rpcProvider`（不导出），通过 `NewProvider()` 返回 interface
- 在 `baseWorker` 上新增 `RegisterProvider(ctx, Provider) error` 方法，参考 `ConnectComponent` 模式
- 在 Worker/Service 的 `Stop()` 流程中统一关闭已注册的 Provider
- 清理 `RpcSender` 中未使用的 `provider` 字段
- 更新 `Worker` interface 以包含 `RegisterProvider` 方法

## Capabilities

### New Capabilities
- `provider-interface`: 定义 Provider interface，将具体实现隐藏为 rpcProvider，使 RPC 调用能力可替换和可 mock
- `register-provider`: 在 Worker/Service 层提供 RegisterProvider 方法，统一管理 Provider 的生命周期注册和停止

### Modified Capabilities

## Impact

- **pkg/rpc/provider/**: Provider 重构为 interface，rpcProvider 成为不导出的具体实现；RpcSender 删除未使用的 provider 字段
- **pkg/runner/types/runner.go**: Worker interface 新增 RegisterProvider 方法
- **pkg/runner/woker.go**: baseWorker 新增 providers 字段、RegisterProvider 和 stopProviders 方法
- **pkg/runner/service.go**: Stop() 流程中加入 stopProviders 调用
- **外部使用方**: `provider.NewProvider()` 返回类型从 `*Provider` 变为 `Provider` interface，调用方式不变
