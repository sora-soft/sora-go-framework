## Why

框架缺少进程级的统一容器。当前 `main()` 需要手动拼装 Logger、Component、Worker、Service 并自行管理生命周期，没有集中的注册和查询机制。需要一个进程唯一的 Runtime 单例，作为框架的全局入口点，持有进程元数据、框架 Logger、以及 Component/Worker/Service 的注册表。

## What Changes

- **新增 `pkg/runtime` 包**：提供进程唯一的 Runtime 单例对象
- **进程元数据**：`startTime`（进程启动时间）、`root`（cwd 快照）
- **框架 Logger**：`FrameLogger`（identify="framework"）、`RpcLogger`（identify="rpc"），通过导出字段让用户自行 `AddOutput`
- **Component 注册表**：`RegisterComponent(name, comp)` 注册到全局名表，`GetComponent(name)` 查询。注册时不 Start，Worker 通过 `GetComponent` + `ConnectComponent` 按需启动
- **Worker/Service 注册表**：`InstallWorker` / `InstallService` 按 id 注册，`UninstallWorker` / `UninstallService` 按 id 移除并调用 `Stop()`
- **Lifecycle 占位**：`Startup(ctx)` / `Shutdown()` 方法暂留空

## Capabilities

### New Capabilities
- `runtime-process-metadata`: 进程级元数据（startTime、root）的获取
- `runtime-logger-management`: 框架 Logger（FrameLogger、RpcLogger）的创建与导出
- `runtime-component-registry`: Component 按 name 注册与查询的全局注册表
- `runtime-worker-service-registry`: Worker/Service 按 id 注册、查询、卸载的注册表

### Modified Capabilities

## Impact

- 新增 `pkg/runtime` 包，无依赖问题（依赖已有的 `pkg/logger`、`pkg/component`、`pkg/runner`）
- Convention 单例模式，无 breaking change
- 不影响现有的 Worker/Service/Component 接口和生命周期
