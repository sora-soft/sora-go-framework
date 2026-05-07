## 上下文

Provider（`pkg/rpc/provider/provider.go`）管理 RPC 发送者的生命周期，通过 watchLoop 监听服务发现端点变化。当前 Start 方法每次调用都会重新创建 context/cancel 并启动新 goroutine，在被多个调用者引用时产生资源泄漏。

component 包的 `baseComponent`（`pkg/component/base.go`）已实现了成熟的引用计数模式，使用 `refCount` + `sync.Mutex` 保护 Start/Stop 的幂等性。Provider 应复用这一模式。

当前 Provider 使用 `sync.RWMutex`（mu）保护 senders map，watchLoop 内部也持有该锁。引用计数的加锁操作与现有锁使用不冲突。

## 目标 / 非目标

**目标：**
- Provider.Start/Stop 支持多次安全调用，通过引用计数配对
- 消除多次 Start 导致的 goroutine 泄漏和 context 泄漏
- 与 component 包的引用计数模式保持一致

**非目标：**
- 不改变 Provider 的 watchLoop 逻辑或 sender 管理策略
- 不改变 RpcSender 的生命周期管理
- 不引入等待 goroutine 退出的同步机制（保持当前的 context 取消模式）

## 决策

### 1. 复用现有 mu（sync.RWMutex）而非引入新 mutex

**选择**：使用 Provider 现有的 `sync.RWMutex` 保护 refCount。

**理由**：component 使用独立的 `sync.Mutex`，但 Provider 已有 mu 保护 senders。Start/Stop 需要在操作 refCount 的同时操作 senders（Stop 时清理），使用同一把锁避免死锁风险。watchLoop 在独立 goroutine 中运行，Start 持锁期间不会与 watchLoop 产生死锁（`go watchLoop()` 立即返回）。

**替代方案**：引入独立的 mutex 专门保护 refCount——增加复杂性，且 Stop 需要同时持有两把锁，引入锁序风险。

### 2. Start 忽略后续 ctx，使用首次传入的 ctx

**选择**：refCount > 0 时，Start 递增 refCount 并返回 nil，不替换 ctx。

**理由**：与 component 模式一致。watchLoop 和所有 RpcSender 已绑定到首次 ctx，替换 ctx 需要重启所有子资源，复杂度远超收益。

### 3. Start/Stop 签名变更为返回 error

**选择**：`Start(ctx context.Context) error`、`Stop() error`。

**理由**：与 component.Component 接口统一。当前 Provider.Start 不返回 error，但未来可能需要（如 discovery 不可用）。Stop 返回 error 与 component 对齐，便于在统一的生命周期管理框架中使用。

## 风险 / 权衡

- **[破坏性 API 变更]** → Start/Stop 签名变更需要所有调用方适配。影响范围有限（Provider 是内部使用），可通过编译错误快速定位。
- **[Stop 不等待 goroutine 退出]** → 保持现有行为：cancel context 后 watchLoop 自行退出，不等待。这是有意的设计选择，避免 Stop 阻塞。
- **[refCount 溢出]** → 理论上 refCount 可溢出 int。实际场景中调用者数量有限，不构成问题。
