### Requirement: Start with reference counting
baseComponent 的 Start 方法 SHALL 使用 mutex 保护 refCount。当 refCount 从 0 变为 1 时，SHALL 调用 impl.Connect(ctx)。当 refCount 已经大于 0 时，SHALL 仅递增 refCount 并立即返回 nil。

#### Scenario: First start triggers connect
- **WHEN** refCount 为 0 时调用 Start(ctx)
- **THEN** refCount SHALL 变为 1，SHALL 调用 impl.Connect(ctx)，ready SHALL 变为 true

#### Scenario: Subsequent start only increments count
- **WHEN** refCount 大于 0 时调用 Start(ctx)
- **THEN** refCount SHALL 递增 1，SHALL NOT 调用 impl.Connect，SHALL 返回 nil

#### Scenario: Connect failure rolls back count
- **WHEN** refCount 为 0 时调用 Start(ctx) 且 impl.Connect(ctx) 返回错误
- **THEN** refCount SHALL 保持为 0，ready SHALL 保持为 false，SHALL 返回该错误

### Requirement: Stop with reference counting
baseComponent 的 Stop 方法 SHALL 使用 mutex 保护 refCount。当 refCount 从 1 变为 0 时，SHALL 调用 impl.Disconnect()。当 refCount 递减后仍大于 0 时，SHALL 仅递减 refCount 并立即返回 nil。

#### Scenario: Last stop triggers disconnect
- **WHEN** refCount 为 1 时调用 Stop()
- **THEN** refCount SHALL 变为 0，SHALL 调用 impl.Disconnect()，ready SHALL 变为 false

#### Scenario: Non-last stop only decrements count
- **WHEN** refCount 大于 1 时调用 Stop()
- **THEN** refCount SHALL 递减 1，SHALL NOT 调用 impl.Disconnect，SHALL 返回 nil

#### Scenario: Stop when already zero
- **WHEN** refCount 为 0 时调用 Stop()
- **THEN** SHALL 返回 nil，SHALL NOT 调用 impl.Disconnect

### Requirement: Concurrent safety
Start 和 Stop 方法 SHALL 在 `sync.Mutex` 保护下执行，确保多个 goroutine 并发调用时 refCount 和 ready 状态的一致性。

#### Scenario: Concurrent start calls
- **WHEN** 两个 goroutine 同时调用同一 baseComponent 的 Start(ctx)
- **THEN** 只有一个 SHALL 调用 impl.Connect(ctx)，另一个 SHALL 仅递增 refCount

#### Scenario: Concurrent start and stop
- **WHEN** 一个 goroutine 调用 Start(ctx) 同时另一个 goroutine 调用 Stop()
- **THEN** 两个操作 SHALL 串行执行，refCount 和 ready 状态 SHALL 保持一致

### Requirement: baseComponent construction
系统 SHALL 提供 `NewBaseComponent(name string, version string, impl)` 构造函数，创建 refCount 为 0、ready 为 false 的 baseComponent 实例。

#### Scenario: New component initial state
- **WHEN** 调用 NewBaseComponent("etcd-main", "0.0.0", impl)
- **THEN** 返回的 baseComponent SHALL Name 为 "etcd-main"，Version 为 "0.0.0"，ready 为 false，refCount 为 0
