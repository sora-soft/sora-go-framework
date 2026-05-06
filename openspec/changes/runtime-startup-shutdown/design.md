## Context

Runtime 是 sora-go-framework 的全局单例注册表，当前位于 `pkg/runner/runtime/` 子包中。它管理 nodeId、components、services、workers 的注册与查找，但 `Startup()` 和 `Shutdown()` 均为空实现。框架缺少统一的生命周期编排入口，消费代码需手动拼装启动/关闭流程。

当前包结构存在 `pkg/runner/runtime` → `pkg/runner` 的单向依赖，若 Runtime.Startup 需要引用 `*NodeRunner`（定义在 `pkg/runner`），则形成循环依赖。将 Runtime 合并到 `pkg/runner` 是解决此问题的最直接方案。

## Goals / Non-Goals

**Goals:**
- 将 Runtime 从 `pkg/runner/runtime/` 合并到 `pkg/runner/`，消除循环依赖障碍
- 实现 `Runtime.Startup(ctx, node, backend)` — 连接 discovery backend，存储 node/backend 引用
- 实现 `Runtime.Shutdown()` — 并发停止所有 worker/service，等待退出后停止 node，最后断开 backend
- 提供并发安全的 `GetNode()`、`GetBackend()`、`GetDiscovery()` 访问方法
- 更新所有受影响的 import 路径

**Non-Goals:**
- Node 注册到 discovery（NodeMetaData → discovery.NodeMeta 转换待后续变更处理）
- 从 discovery 反注册 Node（跟随注册逻辑延后）
- 为 Runtime 添加状态机（LifeCycle），仅做基本的 nil 检查
- 自动注入 Registry 到 Service（保持现有 `SetRegistry` 手动注入方式）
- 改变 Service/Worker 的创建和启动流程

## Decisions

### D1: 包合并而非接口抽象

**选择**: 将 `pkg/runner/runtime/` 整体合并到 `pkg/runner/`。

**备选方案**:
- A) 在 runtime 包定义 Node 接口 → 但返回类型 `NodeMetaData` 仍在 `pkg/runner` 中，仍有循环
- B) Startup 接受 `discovery.NodeMeta` 数据 → 无法存储 `*NodeRunner` 引用
- C) 移动 NodeRunner 到 `pkg/runner/types/` → 改动范围大

**理由**: 方案 A/B 都无法同时满足"存储 NodeRunner 引用"和"无循环依赖"。包合并后所有类型同包可互访，import 变更仅 3 个文件，风险可控。

### D2: Shutdown 并发停止 worker/service

**选择**: 使用 `sync.WaitGroup` 或 `errgroup` 并发调用所有 UninstallService/UninstallWorker，然后停止 node，最后 Disconnect backend。

**备选方案**:
- A) 顺序停止所有 service → 然后停止所有 worker → 然后断开 backend
- B) 完全并发停止所有（含 node）

**理由**: Service 之间无依赖关系，并发停止减少关闭时间。Node 是进程级概念，需在所有 worker/service 退出后才停止，保证关闭期间 node 信息仍可查询。Backend 最后断开，确保 service 反注册 discovery 时 Registry 仍可用。

### D3: node/backend 字段共用一把锁

**选择**: 新增 `nodeMu sync.RWMutex` 保护 `node` 和 `backend` 两个字段。

**理由**: 这两个字段总是在同一时刻（Startup/Shutdown）被同时设置或清除，无独立变更的必要，共用一把锁简化了代码。不复用 `compMu` 避免不必要的锁竞争。

### D4: Startup 不做防重复启动检查

**选择**: 不添加 started 标志位，允许重复调用（由消费代码保证只调用一次）。

**理由**: 当前阶段保持简单。Runtime 作为 framework 级别组件，调用方（消费代码的 main）应保证只调用一次 Startup。若未来需要，可加状态机保护。

## Risks / Trade-offs

- **[Breaking: import 路径]** 所有 `pkg/runner/runtime` 的外部引用需更新 → 仅 3 个文件受影响，改动明确
- **[Breaking: Startup 签名]** `Startup() error` → `Startup(ctx, node, backend) error` → 当前 Startup 是空实现，外部无调用
- **[并发 Shutdown 竞态]** Shutdown 与外部 UninstallService/Worker 并发可能产生冲突 → Shutdown 收集 ID 后逐个调用 Uninstall，Uninstall 对不存在的 ID 返回 nil，安全
- **[Backend.Connect 语义]** 不同 backend 实现（etcd vs ram）的 Connect 行为不同 → 由 backend 实现负责，Runtime 不感知
