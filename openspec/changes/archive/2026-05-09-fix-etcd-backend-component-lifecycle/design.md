## Context

EtcdBackend 当前假设其依赖的 etcd 组件已由调用者提前启动（见 `etcd_test.go:38` 手动调用 `etcdComp.Start(ctx)`）。生产代码中若调用者未启动组件，`Connect()` 只能得到一个模糊的 "not connected" 错误。

baseComponent 已实现引用计数机制（`Start` 递增、`Stop` 递减），多次调用安全。baseWorker.ConnectComponent 也采用了"主动调用 Start"的模式。EtcdBackend 应保持一致。

当前结构体没有保存 Component 引用，无法在 Disconnect 时调用 Stop。

## Goals / Non-Goals

**Goals:**
- EtcdBackend.Connect() 主动调用 component.Start(ctx) 确保组件已连接
- EtcdBackend.Disconnect() 调用 component.Stop() 释放引用
- 清除 initFromEtcd 中的 debug println
- 保持现有测试兼容（测试中的手动 Start 不应破坏，因引用计数幂等）

**Non-Goals:**
- 不改变 Component 接口或 baseComponent 的引用计数机制
- 不引入新的错误类型（复用现有的 newComponentNotFoundError / newComponentTypeError）
- 不修改其他 Backend 实现（如 RAM backend）

## Decisions

### Decision 1: Backend 拥有组件引用生命周期

**选择**: EtcdBackend 在 Connect 中调用 Start，在 Disconnect 中调用 Stop。

**替代方案**: 要求调用者负责启动组件（当前行为）。
- 优点：调用者有完全控制权
- 缺点：容易遗忘，错误信息不明确，与 baseWorker.ConnectComponent 模式不一致

**理由**: baseComponent 的引用计数已处理多消费者场景，Backend 自管理生命周期更健壮。

### Decision 2: 保存 Component 接口引用

**选择**: 在 EtcdBackend 结构体中新增 `comp component.Component` 字段。

**理由**: Disconnect() 需要调用 Stop()，必须持有引用。使用 Component 接口而非具体类型，保持抽象。

### Decision 3: Start 失败时的错误处理

**选择**: Start 失败直接返回错误，不做额外包装。

**理由**: baseComponent.Start 已提供清晰的错误信息（连接失败、选项未设置等），无需额外包装。

## Risks / Trade-offs

- [重复 Start 安全] → baseComponent 引用计数已处理，测试中手动 Start + Backend 内部 Start 等价于 refCount=2，Stop 两次后真正断开。无风险。
- [Disconnect 顺序] → 需确保先清理 lease/watcher（已实现），最后调用 comp.Stop()。如果 Stop 失败，lease/watcher 已清理完毕，不会泄漏。
