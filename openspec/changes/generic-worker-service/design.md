## Context

当前 `baseWorker` 通过嵌入 `types.Runner` 接口持有业务逻辑，`baseService` 嵌入 `*baseWorker`。Runner 的具体类型在构造后丢失，Runner 需要反向引用 Worker/Service 时通过 `WorkerAware`/`ServiceAware` 回调注入。

项目中已有 `BaseComponent[T ComponentImpl]` 泛型模式，成功应用于 etcd 等组件。该模式的核心：Base 持有具体类型 `T`，对外暴露通用接口，对内通过 `.Impl()` 返回具体类型。

Worker/Service 应采用相同模式。

## Goals / Non-Goals

**Goals:**

- `BaseWorker[R Runner]` 泛型化，持有具体 Runner 类型 `R`
- `BaseService[R Runner]` 泛型化，嵌入 `*BaseWorker[R]`，管理 listeners 注册表
- 提供 `.Runner()` 方法返回具体类型 `R`
- 保留 `WorkerAware`/`ServiceAware` 机制，构造函数内部自动注入
- `types.Worker`、`types.Service` 接口不变，Runtime 层无需修改
- 与 `BaseComponent[T]` 保持架构对称

**Non-Goals:**

- 不修改 `types.Worker` / `types.Service` 接口定义
- 不修改 `Runtime` 的注册/管理逻辑
- 不修改 `Component` 层的任何代码
- 不引入 `RunnerImpl` 独立约束接口（直接复用 `Runner`）

## Decisions

### Decision 1: 命名字段 `runner R`，不嵌入类型参数

Go 不允许嵌入泛型类型参数。`BaseWorker[R]` 使用 `runner R` 命名字段持有 Runner，与 `BaseComponent[T]` 的 `impl T` 模式一致。

```
BaseComponent[T]   →  impl T       →  .Impl() T
BaseWorker[R]      →  runner R     →  .Runner() R
```

考虑过：嵌入 `R`（编译错误，Go 不支持）
选择：命名字段 + getter 方法

### Decision 2: BaseService 嵌入 `*BaseWorker[R]`

`BaseService[R]` 通过嵌入 `*BaseWorker[R]` 指针获得 Worker 的全部方法。这允许 `BaseService` 直接满足 `types.Worker` 接口，只需额外实现 `InstallListener`。

### Decision 3: 保留 WorkerAware / ServiceAware

Runner 需要 Service 反向引用（如 `NodeRunner.StateData()` 需要读取 Service 元数据）。构造函数内部通过 `any(runner).(ServiceAware)` 检查并注入，对外接口不变。

考虑过：去掉 Aware 接口
问题：Runner 获取 Service 引用没有其他干净的途径（鸡生蛋问题）
选择：保留 Aware，泛型构造函数内部自动处理

### Decision 4: Listeners 注册表在 BaseService 上

`BaseService.listeners` 作为已安装 listener 的注册表。`InstallListener` 是公开 API，由 Runner 在 `Startup` 中调用。BaseService 不自动安装 listeners——安装时机和数量由 Runner 决定。

正常 Runner：`Startup` 中构造 listener → 调 `svc.InstallListener()`
NodeRunner（特殊）：listeners 外部传入，`Startup` 中遍历安装

### Decision 5: 直接复用 `Runner` 接口作为泛型约束

`Runner` 接口只有 `Startup`/`Shutdown` 两个方法，不需要引入额外的 `RunnerImpl` 约束。

```go
func NewWorker[R Runner](name string, runner R, opts WorkerOptions) *BaseWorker[R]
func NewService[R Runner](name string, runner R, opts ServiceOptions) *BaseService[R]
```

## Risks / Trade-offs

**[Breaking Change] 调用方需指定泛型参数**
→ 影响范围小（仅 `NodeRunner` 和少数业务调用方），迁移是一次性的

**[Go 泛型限制] 无法嵌入类型参数**
→ 使用命名字段，调用处需写 `b.runner.Startup()` 而非 `b.Startup()`
→ 不影响外部接口，仅内部实现细节

**[any() 类型断言] Aware 注入使用 `any(runner).(ServiceAware)`**
→ 与当前模式一致，运行时安全（ok check），无新增风险
