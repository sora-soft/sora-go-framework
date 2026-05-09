## 上下文

当前组件系统由三层构成：

```
Component (接口)              ← 公共 API，runtime 存储用
  └── BaseComponent (结构体)   ← 引用计数、生命周期管理
        └── componentImpl (接口, 包私有)  ← 具体实现约束
```

`BaseComponent` 内嵌 `componentImpl` 接口，`Impl()` 返回 `any`。消费者（如 etcd backend）需要两层类型断言才能调用具体方法。为避免外部重复断言，每种组件都需要一个 `BaseXxxComponent` 包装器（如 `BaseEtcdComponent`），其所有方法都是 `b.Impl().(*ConcreteType).Method()` 的委托。

项目已广泛使用 Go 泛型（`LifeCycle[T]`、`Response[T]`、`watchEntities[T]` 等），泛型基础设施成熟。

## 目标 / 非目标

**目标：**
- 消除所有组件的 `Impl().(*ConcreteType)` 类型断言模式
- 移除 `BaseEtcdComponent` 及未来同类包装器的必要性
- 让新增组件类型的流程从"实现接口 + 写包装器"简化为"实现接口"
- 保持 `Component` 接口不变，runtime 异构存储机制不受影响

**非目标：**
- 不改变组件的引用计数和生命周期管理逻辑
- 不修改 `Component` 接口本身（Start/Stop/LoadOptions/GetMetaInfo）
- 不改变 runtime 的 `map[string]Component` 存储方式
- 不为 `Component` 接口引入泛型参数（避免异构存储问题）

## 决策

### 决策 1：BaseComponent[T ComponentImpl] 泛型化

**选择**：将 `BaseComponent` 改为 `BaseComponent[T ComponentImpl]`，`impl` 字段类型为 `T`。

**理由**：`*BaseComponent[T]` 对任何 T 都满足 `Component` 接口（方法签名不涉及 T），可存入 `map[string]Component`。`Impl()` 返回 `T`，编译期类型安全。

**替代方案**：
- 泛型 `Component[T]` 接口 → 无法存入单个 map（异构类型问题），拒绝
- 保持现状 + 代码生成 → 增加构建复杂度，不解决根本问题，拒绝

### 决策 2：导出 ComponentImpl 接口

**选择**：将 `componentImpl` 重命名为 `ComponentImpl` 并导出。

**理由**：泛型约束需要跨包引用。外部包定义新组件时需实现此接口，包私有约束无法在 `NewBaseComponent[T ComponentImpl]` 中使用。

### 决策 3：impl 字段私有 + Impl() T 公开方法

**选择**：`impl` 字段小写（私有），通过 `Impl() T` 方法公开访问。

**理由**：与 Go 惯例一致（getter 方法），保留未来在 getter 中添加防护逻辑（如 nil 检查、状态验证）的能力。

**替代方案**：
- `Impl` 公开字段 → 更简洁但丧失封装性，拒绝

### 决策 4：双方法模式 — GetComponent + GetComponentOf[T]

**选择**：保留 `GetComponent(name) (Component, bool)` 用于生命周期操作，新增 `GetComponentOf[T](name) (*BaseComponent[T], error)` 用于类型安全访问。

**理由**：`baseWorker.ConnectComponent` 等场景只需要 `Component` 接口，不应被强制指定类型参数。两种用法分离，职责清晰。`GetComponentOf` 返回 `error` 而非 `bool`，区分"未找到"和"类型不匹配"。

**替代方案**：
- 只保留泛型方法 → 每次调用都要指定 T，生命周期场景不友好，拒绝
- 泛型方法返回 bool → 无法区分错误类型，拒绝

### 决策 5：删除 BaseEtcdComponent

**选择**：删除 `BaseEtcdComponent` 结构体及其所有方法。

**理由**：其唯一职责是提供类型安全的委托方法。泛型化后，消费者通过 `comp.Impl()` 直接获得 `*EtcdComponent`，无需中间层。`NewEtcdComponent` 直接返回 `*BaseComponent[*EtcdComponent]`，仍然满足 `Component` 接口。

### 决策 6：etcd 组件内部对 runtime.RT 的引用

**选择**：保持现状，`EtcdComponent` 内部通过 `runtime.RT.FrameLogger` 访问日志器。

**理由**：此依赖与泛型化无关，不在此变更范围内。

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| **BREAKING CHANGE**：`BaseComponent` 变为泛型类型 | 仅影响直接引用 `BaseComponent` 的代码，范围可控（etcd 包 + discovery 包） |
| `GetComponentOf[T]` 内部仍做类型断言，运行时可能失败 | 返回 `error` 明确告知失败原因（未找到 / 类型不匹配），比当前的 `bool` 更友好 |
| Go 泛型类型断言 `c.(*BaseComponent[T])` 的编译器支持 | Go 1.18+ 完全支持，项目已使用泛型 |
| 新增组件开发者需理解泛型约束语法 | 文档化模式：实现 `ComponentImpl` + 调用 `NewBaseComponent[T]`，模板简单 |
