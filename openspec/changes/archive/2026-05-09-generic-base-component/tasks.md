## 1. 核心类型重构

- [x] 1.1 导出 `ComponentImpl` 接口：将 `pkg/component/interface.go` 中的 `componentImpl` 重命名为 `ComponentImpl`（首字母大写），所有方法签名不变
- [x] 1.2 泛型化 `BaseComponent`：将 `pkg/component/base.go` 中的 `BaseComponent` 改为 `BaseComponent[T ComponentImpl]`，`componentImpl` 嵌入字段改为 `impl T` 私有字段，更新所有内部方法中的 `b.componentImpl.Xxx()` 调用为 `b.impl.Xxx()`
- [x] 1.3 更新 `NewBaseComponent` 为泛型构造函数：`func NewBaseComponent[T ComponentImpl](name string, impl T) *BaseComponent[T]`
- [x] 1.4 更新 `Impl()` 方法返回类型为 `T`：`func (b *BaseComponent[T]) Impl() T`，删除原返回 `any` 的实现
- [x] 1.5 更新 `Start`、`Stop`、`LoadOptions`、`GetMetaInfo` 方法的接收者为 `(b *BaseComponent[T])`

## 2. Runtime 泛型方法

- [x] 2.1 在 `pkg/runtime/runtime.go` 中新增包级泛型函数 `GetComponentOf[T component.ComponentImpl](name string) (*component.BaseComponent[T], error)`，内含类型断言和错误区分（Go 不允许非泛型类型上的泛型方法，因此改为包级函数）
- [x] 2.2 确保原有 `GetComponent(name string) (component.Component, bool)` 方法不变

## 3. Etcd 组件适配

- [x] 3.1 更新 `pkg/component/etcd/etcd.go` 中 `NewEtcdComponent` 返回类型为 `*component.BaseComponent[*EtcdComponent]`
- [x] 3.2 删除 `BaseEtcdComponent` 结构体及其全部方法（`Client`、`LeaseID`、`PersistPut`、`PersistDel`、`Keys`、`Lock`、`OnLeaseReconnect`）
- [x] 3.3 删除 `NewBaseEtcdComponent` 构造函数
- [x] 3.4 更新 `EtcdComponent` 内部对 `runtime.RT.FrameLogger` 的引用（如有因类型变更导致的编译问题）

## 4. 消费者代码适配

- [x] 4.1 更新 `pkg/discovery/store/etcd/backend.go`：将 `GetComponent` + 类型断言替换为 `GetComponentOf[*etcdcomp.EtcdComponent]`，通过 `comp.Impl()` 获取具体实现
- [x] 4.2 更新 `pkg/discovery/store/etcd/registry.go`：将 `*etcdcomp.BaseEtcdComponent` 类型引用改为通过 `*component.BaseComponent[*etcdcomp.EtcdComponent]` 或 `*etcdcomp.EtcdComponent`（impl）进行交互

## 5. 验证

- [x] 5.1 编译通过：`go build ./...`
- [x] 5.2 现有测试通过：`go test ./...`
- [x] 5.3 确认无代码引用已删除的 `BaseEtcdComponent` 和 `componentImpl`（包私有名）
