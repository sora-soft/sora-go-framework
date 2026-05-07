## 新增需求

### 需求:引用计数 Start

Provider.Start 必须支持多次安全调用。首次调用必须执行实际的启动逻辑（创建 context、启动 watchLoop），后续调用必须仅递增引用计数并立即返回 nil，忽略新传入的 context。

#### 场景:首次 Start 调用
- **当** Provider 的 refCount 为 0 时调用 Start(ctx)
- **那么** Provider 必须创建新的 context/cancel，启动 watchLoop goroutine，设置 refCount 为 1，并返回 nil

#### 场景:多次 Start 调用
- **当** Provider 的 refCount > 0 时再次调用 Start(newCtx)
- **那么** Provider 必须仅递增 refCount，返回 nil，禁止替换现有 context 或启动新的 watchLoop

#### 场景:Start 线程安全
- **当** 多个 goroutine 并发调用 Start
- **那么** refCount 必须通过互斥锁保护，确保计数准确且不会重复启动

### 需求:引用计数 Stop

Provider.Stop 必须与 Start 配对调用。每次 Stop 必须递减引用计数，仅当计数降至 0 时执行实际的清理逻辑（取消 context、销毁所有 sender）。

#### 场景:Stop 递减但未清零
- **当** refCount > 1 时调用 Stop
- **那么** Provider 必须递减 refCount 并返回 nil，禁止取消 context 或销毁 sender

#### 场景:Stop 清零触发清理
- **当** refCount 为 1 时调用 Stop
- **那么** Provider 必须递减 refCount 至 0，取消 context，销毁所有 sender，并返回 nil

#### 场景:Stop 在未 Start 时调用
- **当** refCount 为 0 时调用 Stop
- **那么** Provider 必须直接返回 nil，禁止执行任何清理操作

#### 场景:Stop 线程安全
- **当** 多个 goroutine 并发调用 Stop
- **那么** refCount 必须通过互斥锁保护，确保清理逻辑仅执行一次

### 需求:方法签名返回 error

Provider.Start 和 Provider.Stop 必须返回 error 类型，与 component.Component 接口保持一致。

#### 场景:Start 返回 error
- **当** 调用 Provider.Start(ctx)
- **那么** 方法签名必须为 `Start(ctx context.Context) error`

#### 场景:Stop 返回 error
- **当** 调用 Provider.Stop()
- **那么** 方法签名必须为 `Stop() error`
