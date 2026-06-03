## Why

`baseWorker` 和 `baseService` 通过接口 `types.Runner` 持有业务逻辑，导致具体 Runner 类型信息丢失。`BaseComponent[T ComponentImpl]` 已成功证明了泛型包装模式的价值——对外暴露通用接口，对内保留具体类型。Worker/Service 应采用相同的模式，统一架构风格并获得类型安全。

## What Changes

- **`baseWorker` → `BaseWorker[R Runner]`**：泛型化，持有具体 Runner 类型 `R`，通过 `Runner() R` 返回具体类型
- **`baseService` → `BaseService[R Runner]`**：泛型化，嵌入 `*BaseWorker[R]`，管理 listeners 注册表
- **`NewWorker` / `NewService` 构造函数泛型化**：`NewWorker[R Runner](...)` / `NewService[R Runner](...)`
- **保留 `WorkerAware` / `ServiceAware`**：Runner 仍需要 Service 反向引用，构造函数内部通过 Aware 接口注入
- **BREAKING**：所有使用 `NewWorker` / `NewService` 的调用方需指定泛型参数

## Capabilities

### New Capabilities

- `generic-worker`: 泛型化的 Worker/Service 基础结构，支持具体 Runner 类型安全访问

### Modified Capabilities

（无现有 specs）

## Impact

- `pkg/runner/woker.go`：重写为 `BaseWorker[R Runner]`
- `pkg/runner/service.go`：重写为 `BaseService[R Runner]`
- `pkg/runner/types/runner.go`：Runner 接口不变，WorkerAware/ServiceAware 保留
- `pkg/node/node.go`：NodeRunner 适配新的泛型构造函数
- `pkg/runtime/runtime.go`：Runtime 持有 `types.Worker`/`types.Service` 接口不变，无需修改
