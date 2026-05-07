## 1. Provider Interface 定义

- [x] 1.1 创建 `pkg/rpc/provider/interface.go`，定义 Provider interface（Start、Stop、CallRpc）
- [x] 1.2 将 `pkg/rpc/provider/provider.go` 中的 `Provider` struct 重命名为 `rpcProvider`（不导出），更新所有方法接收者
- [x] 1.3 更新 `NewProvider()` 返回类型为 `Provider` interface

## 2. RpcSender 清理

- [x] 2.1 从 `RpcSender` struct 中删除 `provider` 字段
- [x] 2.2 更新 `NewRpcSender` 签名，移除 `provider` 参数

## 3. Runner 层集成

- [x] 3.1 在 `baseWorker` 中新增 `providers []provider.Provider` 字段和 `provMu sync.Mutex`
- [x] 3.2 实现 `RegisterProvider(ctx context.Context, p provider.Provider) error` 方法
- [x] 3.3 实现 `stopProviders()` 方法
- [x] 3.4 在 `baseWorker.Stop()` 中 `Shutdown` 之后、`disconnectComponents` 之前调用 `stopProviders()`
- [x] 3.5 在 `baseService.Stop()` 中 `Shutdown` 之后、`disconnectComponents` 之前调用 `stopProviders()`

## 4. Interface 更新

- [x] 4.1 在 `types.Worker` interface 中新增 `RegisterProvider(ctx context.Context, p provider.Provider) error`

## 5. 验证

- [x] 5.1 更新 `cmd/sora-test/main.go` 中对 Provider 的使用，确保兼容 interface 返回类型
- [x] 5.2 编译通过，无编译错误
