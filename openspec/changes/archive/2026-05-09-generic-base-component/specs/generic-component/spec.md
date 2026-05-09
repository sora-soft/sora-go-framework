## 新增需求

### 需求:泛型 BaseComponent 定义
`BaseComponent` 必须（MUST）定义为泛型结构体 `BaseComponent[T ComponentImpl]`，其中类型参数 `T` 必须满足 `ComponentImpl` 接口约束。结构体必须（MUST）包含私有的 `impl` 字段（类型为 `T`），并通过公开方法 `Impl() T` 提供访问。

#### 场景:创建泛型 BaseComponent 实例
- **当** 调用 `NewBaseComponent[T](name, impl)` 且 `impl` 满足 `ComponentImpl` 接口
- **那么** 返回 `*BaseComponent[T]` 实例，其 `Impl()` 方法返回与传入参数类型完全相同的 `T`

#### 场景:Impl 方法返回具体类型
- **当** 对 `*BaseComponent[*EtcdComponent]` 调用 `Impl()`
- **那么** 返回值类型为 `*EtcdComponent`，无需类型断言

### 需求:ComponentImpl 接口导出
原 `componentImpl`（包私有）必须（MUST）重命名为 `ComponentImpl` 并导出。接口方法签名不得（MUST NOT）变更：`Connect`、`Disconnect`、`SetOptions`、`GetOptions`、`GetVersion`。

#### 场景:跨包实现 ComponentImpl
- **当** 外部包定义结构体并实现 `ComponentImpl` 的全部方法
- **那么** 该结构体可作为 `NewBaseComponent[T]` 的类型参数和 `impl` 参数

### 需求:BaseComponent[T] 满足 Component 接口
`*BaseComponent[T]` 对任何满足 `ComponentImpl` 的 T 必须（MUST）自动满足 `Component` 接口。`Start`、`Stop`、`LoadOptions`、`GetMetaInfo` 方法签名不得（MUST NOT）涉及类型参数 T。

#### 场景:存入 runtime 组件注册表
- **当** 创建 `*BaseComponent[*EtcdComponent]` 并调用 `RegisterComponent(name, comp)`
- **那么** 组件成功存入 `map[string]Component`，后续可通过 `GetComponent` 检索

#### 场景:引用计数行为不变
- **当** 对同一个 `*BaseComponent[T]` 多次调用 `Start`
- **那么** 仅第一次调用触发 `impl.Connect`，`refCount` 递增；`Stop` 递减 `refCount`，仅在归零时调用 `impl.Disconnect`

### 需求:泛型 GetComponentOf 方法
Runtime 必须（MUST）提供泛型方法 `GetComponentOf[T ComponentImpl](name string) (*BaseComponent[T], error)`。此方法必须（MUST）在内部完成从 `Component` 到 `*BaseComponent[T]` 的类型断言，并对调用者隐藏此细节。

#### 场景:成功获取类型安全的组件
- **当** 调用 `GetComponentOf[*EtcdComponent]("etcd")` 且注册表中存在类型匹配的组件
- **那么** 返回 `*BaseComponent[*EtcdComponent]` 和 `nil` error

#### 场景:组件不存在
- **当** 调用 `GetComponentOf[T]("unknown")` 且名称不在注册表中
- **那么** 返回 `nil` 和错误信息，包含组件名称

#### 场景:类型不匹配
- **当** 调用 `GetComponentOf[*RedisComponent]("etcd")` 但注册的组件 impl 类型为 `*EtcdComponent`
- **那么** 返回 `nil` 和类型不匹配错误

### 需求:保留非泛型 GetComponent 方法
`GetComponent(name string) (Component, bool)` 必须（MUST）继续存在，用于只需要 `Component` 接口（Start/Stop/GetMetaInfo）的场景。返回类型和语义不得（MUST NOT）变更。

#### 场景:仅生命周期操作
- **当** 调用 `GetComponent("etcd")` 获取组件并仅调用 `Start`/`Stop`
- **那么** 正常工作，与变更前行为一致

### 需求:消除 BaseEtcdComponent 包装器
`BaseEtcdComponent` 结构体及其所有方法必须（MUST）被删除。`NewEtcdComponent` 构造函数必须（MUST）直接返回 `*BaseComponent[*EtcdComponent]`。消费者必须（MUST）通过 `comp.Impl()` 获取 `*EtcdComponent` 实例并直接调用其方法。

#### 场景:消费者访问 etcd 特定方法
- **当** 通过 `GetComponentOf[*EtcdComponent]` 获取组件
- **那么** 调用 `comp.Impl().Client()` 直接获得 `*clientv3.Client`，无需任何类型断言

#### 场景:新增组件类型无需编写包装器
- **当** 新增 `RedisComponent` 实现 `ComponentImpl` 接口
- **那么** 只需调用 `NewBaseComponent[*RedisComponent](name, impl)` 即可，无需编写 `BaseRedisComponent` 包装器

## 修改需求

## 移除需求
