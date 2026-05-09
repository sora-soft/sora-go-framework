## 为什么

`BaseComponent` 的 `Impl()` 方法返回 `any`，导致消费者必须进行类型断言才能访问具体实现的方法。当前 etcd 组件通过 `BaseEtcdComponent` 包装器解决此问题，但该包装器的每个方法都重复 `b.Impl().(*EtcdComponent)` 这一模式——7 个方法，7 次相同的断言。随着未来更多组件类型（Redis、MQ 等）的加入，每新增一种组件都要编写对应的 BaseXxxComponent 包装器，产生大量样板代码。Go 1.18+ 的泛型可以彻底消除这个问题。

## 变更内容

- **泛型化 `BaseComponent`**：将 `BaseComponent` 改为 `BaseComponent[T ComponentImpl]`，`impl` 字段类型从接口变为具体类型 `T`，`Impl()` 方法返回 `T` 而非 `any`
- **导出 `ComponentImpl` 接口**：原 `componentImpl`（包私有）导出为 `ComponentImpl`，作为泛型约束
- **新增 `GetComponentOf[T]`**：在 Runtime 上新增泛型方法，内部完成类型断言，对外返回 `*BaseComponent[T]`
- **删除 `BaseEtcdComponent`**：该包装器不再需要，消费者直接通过 `comp.Impl()` 获取 `*EtcdComponent`
- **BREAKING**：`BaseComponent` 从具体类型变为泛型类型，所有直接引用 `BaseComponent` 的代码需更新为 `BaseComponent[T]`
- **BREAKING**：`componentImpl` 接口重命名为 `ComponentImpl`（导出）
- **BREAKING**：`NewBaseComponent` 签名从 `func NewBaseComponent(name string, impl componentImpl) *BaseComponent` 变为 `func NewBaseComponent[T ComponentImpl](name string, impl T) *BaseComponent[T]`

## 功能 (Capabilities)

### 新增功能
- `generic-component`: 泛型组件基础设施，包括泛型 BaseComponent[T]、导出的 ComponentImpl 接口约束、泛型 GetComponentOf[T] 方法

### 修改功能

## 影响

- **pkg/component/**：`interface.go`（导出接口）、`base.go`（泛型化）
- **pkg/component/etcd/**：`etcd.go`（删除 BaseEtcdComponent，简化构造函数）
- **pkg/runtime/**：`runtime.go`（新增 GetComponentOf[T] 方法）
- **pkg/discovery/store/etcd/**：`backend.go`、`registry.go`（改用 GetComponentOf，消除类型断言）
- **API 变更**：所有使用 `BaseComponent` 的包需适配泛型语法
