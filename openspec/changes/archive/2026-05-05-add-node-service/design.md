## Context

`pkg/runner` 已建立三层设计模式（Runner 接口 / Worker+Service 接口 / baseWorker+baseService 实现），`pkg/runtime` 作为全局单例管理 Component/Worker/Service 注册表。但 `pkg/runtime` import `pkg/runner`（使用 Worker、Service 类型），而 NewWorker 又需要访问 Runtime 的 nodeId——形成了 `runner ↔ runtime` 循环依赖。

此外，框架缺少"节点"概念。TS 参考实现中 `Node extends Service`，在 Startup 时安装 Listeners，并提供 `nodeStateData`（节点自身元数据）和 `nodeRunData`（进程快照）。当前 Go 实现中 `WorkerMetaData.NodeId` 恒为空字符串，没有节点身份传播机制。

## Goals / Non-Goals

**Goals:**
- 提取 `pkg/runner/types` 子包，打破 runner ↔ runtime 循环依赖
- 迁移 `pkg/runtime` → `pkg/runner/runtime`，使 NewWorker 可直接访问 RT.NodeId()
- 实现 NodeRunner（Runner + ServiceRefAware），在 Startup 中安装 Listeners 并注册 NodeId
- 定义 NodeMetaData / NodeRunData 结构体，提供节点状态查询和进程快照
- NewWorker 构造时自动从 RT.NodeId() 填充 WorkerMetaData.NodeId

**Non-Goals:**
- 不实现 Discovery / Scope（TS 中的 `Runtime.discovery`、`Runtime.scope`），后续按需添加
- 不实现 Worker/Service 的工厂注册表（registerWorker/registerService + factory），Go 中直接构造即可
- 不实现动态添加/移除 Listener 的运行时管理
- 不实现节点间通信或集群发现

## Decisions

### D1: 三子包结构（types / runtime / root）

将 `pkg/runner` 拆为三层：`runner/types`（接口 + 元数据类型）、`runner/runtime`（Runtime 单例）、`runner` 根包（实现）。

**理由**: Go 包是唯一封装边界。TypeScript 的 class 既有实例方法也有静态方法，共享类型作用域；Go 需要显式的子包来解耦。这是 Go 大型项目（k8s、grpc-go、etcd）的惯用做法。

**依赖方向**: `runner` → `runner/types`, `runner/runtime`；`runner/runtime` → `runner/types`。无环。

**替代方案**: 合并到同一包（runner 包职责混杂）、包级变量 hack（全局可变状态）、显式传参（NewWorker 签名膨胀）。

### D2: NodeRunner 实现 Runner + ServiceRefAware

NodeRunner 实现 Runner 接口（Startup/Shutdown），同时实现 ServiceRefAware 接口以获取 ServiceRef。在 Startup 中通过 ServiceRef.InstallListener 安装所有预配置的 Listeners，然后调用 runtime.RT.SetNodeId 注册自身 ID。

**理由**: 与现有 runner 三层设计一致——消费者通过 Service 接口操作，实现者只需关心 Runner 接口。ServiceRefAware 注入让 NodeRunner 能在 Startup 中使用框架提供的 Listener 管理能力。

### D3: NodeId 通过 Runtime 单例传播

Runtime 存储 nodeId 字段。NodeRunner.Startup 时调用 RT.SetNodeId(svc.GetMetadata().Id)。NewWorker 构造时调用 RT.NodeId() 读取并填入 WorkerMetaData.NodeId。

**理由**: 最简单的传播方式。Go 不支持类继承和静态属性，全局单例是等价方案。NodeId 在进程生命周期内只设置一次（Startup），不存在并发写入问题。

### D4: Version 硬编码 "0.0.0"

NodeVersions.Framework 和 App 均硬编码为 "0.0.0"，后续通过构建时注入（ldflags）或配置文件替换。

**理由**: 当前阶段不需要版本管理基础设施，硬编码避免过度设计。

### D5: NodeMetaData 包含进程级信息

NodeMetaData 包含 Host（os.Hostname()）、Pid（os.Getpid()）、Versions，这些在 NodeRunner 构造时采集。

**理由**: 对齐 TS 参考实现。进程级元数据在节点启动时确定，生命周期内不变。

### D6: NodeRunData 从 Runtime 注册表聚合

NodeRunData 聚合 RT 中的 Components、Services、Workers 元数据加上 NodeMetaData，提供完整的进程快照。

**理由**: 与 TS 的 `nodeRunData` getter 对齐。所有数据源都在 Runtime 中，NodeRunner 只是聚合查询。

## Risks / Trade-offs

- **[包结构迁移影响面大]** → `pkg/runtime` → `pkg/runner/runtime`、接口迁移至 `runner/types`，所有 import 路径变更。通过 tasks 中明确的迁移步骤控制风险。
- **[runner/types 增加一层间接]** → 消费者需 import 两个包（`runner/types` + `runner` 或 `runner/runtime`）。但这是 Go 解循环依赖的标准代价。
- **[RT.NodeId 全局可变状态]** → 理论上 SetNodeId 可被多次调用。但语义上只在 NodeRunner.Startup 时调用一次，且单进程只有一个 Node。风险可接受。
- **[NodeRunner 持有 Service 引用]** → NodeRunner 需要知道 Service 的 ID 来设置 NodeId，但 Service 在 NewService 构造时才生成 ID。Startup 时 ServiceRef 是 baseService 指针，ID 已生成。时序正确。
