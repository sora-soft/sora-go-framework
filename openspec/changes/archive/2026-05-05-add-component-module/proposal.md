## Why

框架当前缺少对外部基础设施（etcd、MySQL、Redis 等）的统一连接管理抽象。每个外部组件的连接/断开逻辑需要各自重复实现，无法支持多个 Worker 共享同一组件实例时的引用计数管理。需要一个通用的 Component 模块来标准化外部组件的生命周期管理。

## What Changes

- 新增 `pkg/component` 包，提供外部组件连接管理的抽象层
- 定义 `Component` 对外接口（Start、Stop、LoadOptions、GetMetaInfo）
- 定义内部实现接口（Connect、Disconnect、SetOptions、GetOptions）
- 实现 `baseComponent` 基类，提供引用计数机制：refCount 从 0→1 时调用 connect，从 1→0 时调用 disconnect
- 使用 mutex 全保护确保并发安全
- connect 失败时 refCount 回退到 0

## Capabilities

### New Capabilities
- `component-interface`: Component 对外接口定义与 ComponentMetadata 类型
- `component-lifecycle`: baseComponent 引用计数生命周期管理（Start/Stop/connect/disconnect）
- `component-options`: LoadOptions 与 SetOptions 选项加载机制
- `component-metadata`: GetMetaInfo 元数据暴露（name、ready、version、options）

### Modified Capabilities

## Impact

- 新增 `pkg/component/` 包，不影响现有代码
- 后续 `baseWorker` 将新增 `connectComponent`/`disconnectComponent` 方法来集成 Component（本次变更范围外，由 Runtime 模块负责）
- 依赖现有的 `pkg/utility/errorx` 进行错误处理
