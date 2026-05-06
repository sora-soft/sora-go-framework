## Why

`pkg/runner` 的 `baseWorker` 和 `BaseService` 当前直接暴露 struct，没有公共接口层，与 `pkg/component` 的三层设计（`Component` / `componentImpl` / `baseComponent`）不一致。同时缺少 `connectComponent` / `installListener` 机制，外部 Runner 无法在 Startup 中将创建的组件和 Listener 注册给框架管理生命周期。

## What Changes

- **新增 `Worker` 公共接口**：定义 `Start`、`Stop`、`Go`、`GetMetadata`，消费者通过接口操作
- **新增 `Service` 公共接口**：嵌入 `Worker`，语义上表示 Service is a Worker
- **新增 `WorkerRef` / `ServiceRef` 接口**：提供 `ConnectComponent` / `InstallListener` 注册方法，立即生效
- **新增 `WorkerRefAware` / `ServiceRefAware` 接口**：Runner 可选实现，构造时自动注入 Ref
- **`baseWorker` 改为 unexported**：实现 `Worker` + `WorkerRef`，新增 `components` 切片和 `connectComponent` / `disconnectComponents` 方法
- **`BaseService` 重命名为 `baseService`（unexported）**：实现 `Service` + `ServiceRef`，新增 `listeners` 切片和 `installListener` / `stopListeners` 方法
- **`WorkerMetaData` 统一结构**：合并 Worker 和 Service 的元数据字段（Labels、Listeners），Worker 留空，Service 填充
- **`NewWorker` 返回 `Worker` 接口，`NewService` 返回 `Service` 接口**
- **Startup 失败时自动回滚**：断开已连接的 Component，进入 Error 状态
- **Service.Stop override Worker.Stop**：先停 Listener 再走 Worker 关停流程

## Capabilities

### New Capabilities
- `runner-interface`: Worker / Service 公共接口定义，以及 Runner hook 接口
- `runner-ref-injection`: WorkerRef / ServiceRef 注入机制，Runner 可选实现 RefAware 接口获取框架能力
- `worker-component-management`: baseWorker 的 Component 注册、连接、断开管理
- `service-listener-management`: baseService 的 Listener 安装、停止管理

### Modified Capabilities
- `worker-lifecycle`: Start 流程增加 Startup 失败时自动 disconnectComponents 回滚；GetMetadata 返回统一大结构；baseWorker 改为 unexported

## Impact

- **BREAKING**: `NewWorker` 返回类型从 `*baseWorker` 改为 `Worker` 接口
- **BREAKING**: `NewService` 需新增构造器，返回 `Service` 接口
- **BREAKING**: `baseWorker` 改为 unexported，外部无法直接引用 struct
- **BREAKING**: `BaseService` 重命名为 unexported `baseService`
- **BREAKING**: `WorkerMetaData` 新增 `Labels` 和 `Listeners` 字段
- `pkg/runner/interface.go`：新增 Worker、Service、WorkerRef、ServiceRef、WorkerRefAware、ServiceRefAware 接口
- `pkg/runner/worker.go`：重构 baseWorker，新增 components 管理
- `pkg/runner/service.go`：重构为 baseService，新增 listeners 管理、Stop override
- 依赖：`pkg/component` 的 `Component` 接口、`pkg/rpc` 的 `Listener` 类型
