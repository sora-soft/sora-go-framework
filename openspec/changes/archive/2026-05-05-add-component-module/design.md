## Context

sora-go-framework 是一个 Go 微服务框架，当前已有 `runner` 包管理 Worker/Service 的生命周期，`rpc` 包管理网络连接。框架缺少对外部基础设施（etcd、MySQL、Redis 等）的统一连接管理。当前各外部组件需要各自实现连接逻辑，无法支持多 Worker 共享同一组件实例的场景。

## Goals / Non-Goals

**Goals:**
- 提供通用的 Component 抽象，标准化外部组件的连接管理
- 通过引用计数支持多个消费者共享同一组件实例
- 并发安全地管理 Start/Stop 调用
- 暴露统一的元数据接口（name、ready、version、options）
- 为后续 Runtime 单例模块提供组件管理基础

**Non-Goals:**
- 不实现具体组件（etcd、MySQL 等），只提供基类和接口
- 不实现 Runtime 单例的组件注册/创建逻辑
- 不实现 baseWorker 与 Component 的集成（connectComponent/disconnectComponent）
- 不实现组件间依赖管理
- 不实现 options 的 merge/diff，SetOptions 为直接赋值

## Decisions

### D1: 双层接口设计

对外接口 `Component` 暴露给消费者，包含 Start/Stop/LoadOptions/GetMetaInfo。内部接口无命名，具体组件实现 Connect/Disconnect/SetOptions/GetOptions。baseComponent 持有内部接口实例，实现对外接口。

**理由**: 消费者只需关心 Start/Stop，无需了解 connect/disconnect 的细节。引用计数逻辑封装在 baseComponent 内部。

### D2: mutex 全保护而非 atomic 快路径

Start 和 Stop 使用 `sync.Mutex` 保护 refCount 和 ready 状态。connect/disconnect 在锁内执行。

**理由**: 虽然理论上 atomic 快路径性能更好，但存在竞态窗口——并发 Start 时先到的 connect 失败会导致后到的调用者拿到无效引用。mutex 全保护消除所有竞态，且 connect/disconnect 调用频率低，mutex 开销可忽略。

**替代方案**: atomic.Add + mutex + ready 等待（sync.Cond）。因实现复杂度显著上升且收益有限，未采纳。

### D3: connect 失败时 refCount 回退到 0

当 refCount 从 0→1 时调用 connect，如果 connect 失败，refCount 保持为 0，ready 保持为 false。

**理由**: 保证调用者收到的错误意味着组件未启动，后续可重试 Start。

### D4: GetMetaInfo 由 baseComponent 实现

元数据（name、ready、version、options）由 baseComponent 从自身字段 + impl.GetOptions() 组装，具体组件无需实现。

**理由**: name、version、ready 是 baseComponent 的管理职责，options 通过 impl.GetOptions() 获取。具体组件只需实现 4 个方法（Connect/Disconnect/SetOptions/GetOptions）。

### D5: 版本号硬编码为 "0.0.0"

当前阶段 Component 的 version 字段硬编码为 "0.0.0"，由 NewBaseComponent 参数传入。

**理由**: 未来可根据需要改为从构建信息注入，当前无需增加复杂度。

## Risks / Trade-offs

- **[锁内 connect 阻塞]** → connect 在 mutex 内执行，如果外部组件连接超时（如 etcd 30s），会阻塞其他 Start/Stop 调用。缓解：调用方应使用带 timeout 的 context。
- **[无 LifeCycle 状态机]** → Component 只有 ready bool，没有像 Worker 那样的多状态 LifeCycle。如果未来需要更细粒度的状态（如 connecting、reconnecting），需要引入 LifeCycle。当前权衡：保持简单，ready bool 足够。
- **[Options 类型为 any]** → LoadOptions/SetOptions/GetOptions 使用 any 类型，类型安全由具体组件保证。这是 Go 泛型在该场景下的权衡——对基类来说 any 足够。
