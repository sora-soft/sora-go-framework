## ADDED Requirements

### 需求:baseWorker 持有 Provider 列表
`baseWorker` 必须持有 `[]provider.Provider` 字段和对应的互斥锁（`provMu sync.Mutex`）来管理已注册的 Provider。

#### 场景:字段存在
- **当** baseWorker 初始化
- **那么** providers 列表为空，provMu 为未锁定的 sync.Mutex

### 需求:RegisterProvider 方法
`baseWorker` 必须提供 `RegisterProvider(ctx context.Context, p provider.Provider) error` 方法。该方法必须：
1. 调用 `p.Start(ctx)`
2. 如果 Start 返回错误，立即返回该错误
3. 如果 Start 成功，将 p 追加到 providers 列表（加锁保护）

#### 场景:成功注册
- **当** 调用 RegisterProvider 且 Provider.Start 成功
- **那么** Provider 被追加到 providers 列表，方法返回 nil

#### 场景:Start 失败
- **当** 调用 RegisterProvider 且 Provider.Start 返回错误
- **那么** Provider 不被追加到列表，方法返回该错误

#### 场景:并发注册
- **当** 多个 goroutine 同时调用 RegisterProvider
- **那么** providers 列表的修改必须受到互斥锁保护，不会产生数据竞争

### 需求:stopProviders 方法
`baseWorker` 必须提供 `stopProviders()` 方法，遍历所有已注册的 Provider 并调用其 `Stop()` 方法。

#### 场景:停止所有 Provider
- **当** stopProviders 被调用
- **那么** 所有已注册的 Provider 的 Stop 方法被依次调用

### 需求:Worker Stop 流程集成
Worker 的 `Stop()` 方法必须在 `Shutdown()` 之后、`disconnectComponents()` 之前调用 `stopProviders()`。

#### 场景:Worker 停止顺序
- **当** Worker.Stop() 被调用
- **那么** 执行顺序为：cancel → wg.Wait → Shutdown → stopProviders → disconnectComponents → Stopped

### 需求:Service Stop 流程集成
Service 的 `Stop()` 方法必须在 `Shutdown()` 之后、`disconnectComponents()` 之前调用 `stopProviders()`。

#### 场景:Service 停止顺序
- **当** Service.Stop() 被调用
- **那么** 执行顺序为：stopListeners → cancel → wg.Wait → Shutdown → stopProviders → disconnectComponents → Stopped

### 需求:Worker interface 更新
`types.Worker` interface 必须新增 `RegisterProvider(ctx context.Context, p provider.Provider) error` 方法。

#### 场景:interface 满足
- **当** baseWorker 实现了 RegisterProvider 方法
- **那么** baseWorker 必须满足更新后的 Worker interface

### 需求:Service 通过嵌入继承
由于 `baseService` 嵌入了 `baseWorker`，`RegisterProvider` 方法必须自动对 Service 可用，无需在 baseService 上单独实现。

#### 场景:Service 调用 RegisterProvider
- **当** 通过 Service 实例调用 RegisterProvider
- **那么** 调用实际由嵌入的 baseWorker 处理，行为与 Worker 一致
