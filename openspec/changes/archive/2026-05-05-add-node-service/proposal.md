## Why

框架缺少进程级的 Node 概念。当前 `runtime.RT` 管理了 Component/Worker/Service 注册表，但没有"节点自身"的身份和状态表达。需要一个 NodeRunner 作为顶层 Runner，在 Startup 时安装 Listeners、向 Runtime 注册自身 NodeId，并为后续 Worker/Service 提供 NodeId 传播能力。同时需要重构包结构，将 `pkg/runtime` 移入 `pkg/runner/runtime` 并提取 `pkg/runner/types` 子包，以解决 runner ↔ runtime 的循环依赖问题。

## What Changes

- **新增 `pkg/runner/types` 子包**：将 Worker、Service、Runner、WorkerRef、ServiceRef、WorkerRefAware、ServiceRefAware 接口及 WorkerMetaData、WorkerState 等类型提取到独立子包，打破 runner ↔ runtime 循环依赖
- **迁移 `pkg/runtime` → `pkg/runner/runtime`**：Runtime 单例成为 runner 的子包，新增 `nodeId` 字段及 `NodeId()` / `SetNodeId()` 方法，存储的注册表类型改为引用 `types` 子包
- **新增 NodeRunner**：实现 `types.Runner` + `types.ServiceRefAware` 接口，在 `Startup` 中通过 `svcRef.InstallListener` 安装所有 Listeners，调用 `runtime.RT.SetNodeId()` 注册 NodeId；提供 `StateData()` 返回 `NodeMetaData`、`RunData()` 返回 `NodeRunData` 进程快照
- **新增 NodeMetaData / NodeRunData / NodeVersions 结构体**：NodeMetaData 包含 Id、Alias、Host、Pid、Endpoints、State、StartTime、Versions；NodeRunData 聚合 Node、Components、Services、Workers 元数据
- **NewWorker 自动填充 NodeId**：`NewWorker()` 构造时从 `runtime.RT.NodeId()` 读取当前节点 ID 写入 `WorkerMetaData.NodeId`

## Capabilities

### New Capabilities
- `node-runner`: NodeRunner 实现 Runner + ServiceRefAware，负责在 Startup 中安装 Listeners、注册 NodeId，提供 StateData/RunData 查询
- `node-metadata`: NodeMetaData / NodeRunData / NodeVersions 结构体定义，描述节点自身状态和进程级快照

### Modified Capabilities
- `worker-lifecycle`: NewWorker 构造时从 runtime.RT.NodeId() 自动填充 WorkerMetaData.NodeId 字段（原为空字符串）
- `runner-interface`: 接口和元数据类型迁移至 `pkg/runner/types` 子包，原有 `pkg/runner` 导出类型路径变更

## Impact

- **BREAKING**: `pkg/runtime` 包迁移至 `pkg/runner/runtime`，所有 `import "pkg/runtime"` 需改为 `import "pkg/runner/runtime"`
- **BREAKING**: `runner.Worker`、`runner.Service`、`runner.Runner` 等接口类型迁移至 `runner/types` 子包，import 路径变更
- **BREAKING**: `WorkerMetaData`、`WorkerState` 等类型迁移至 `runner/types` 子包
- 新增 `pkg/runner/types/` 子包
- 新增 `pkg/runner/runtime/` 子包（替代原 `pkg/runtime/`）
- 删除 `pkg/runtime/` 包
- `pkg/runner/woker.go`、`pkg/runner/service.go` 的 import 路径调整
