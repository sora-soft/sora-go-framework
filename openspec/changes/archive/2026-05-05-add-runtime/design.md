## Context

当前框架缺少进程级容器。`cmd/sora-test/main.go` 中所有 Logger、Component、Transport、Listener 的创建和生命周期管理都散落在 `main()` 中，没有统一的注册和查询机制。

现有基础：
- `pkg/component`：已实现 Component 接口 + baseComponent（含 refCount 共享机制）
- `pkg/runner`：已实现 Worker/Service 公共接口 + Ref 注入机制（ConnectComponent / InstallListener）
- `pkg/logger`：已实现 Logger（identify + 多 Output）

需要 Runtime 作为顶层容器，将零散的手动拼装变成声明式注册。

## Goals / Non-Goals

**Goals:**
- 提供进程唯一的 Runtime 单例，作为框架的全局入口点
- 持有进程元数据（startTime、root）
- 创建并暴露 FrameLogger（"framework"）和 RpcLogger（"rpc"），用户自行 AddOutput
- 提供 Component 按 name 的全局注册表（注册不启动，Worker 按需 ConnectComponent）
- 提供 Worker/Service 按 id 的注册表，卸载时自动 Stop
- Startup/Shutdown 占位，为未来生命周期编排预留

**Non-Goals:**
- 不实现 Runtime 对 Component 的生命周期管理（Start/Stop 由 Worker 的 ConnectComponent 负责）
- 不实现 Worker/Service 的依赖管理或启动顺序编排
- 不实现 Runtime 自身的状态机（当前为简单 struct，不做 Lifecycle 管理）
- 不实现配置文件加载或环境变量读取

## Decisions

### D1: Convention 单例（非 sync.Once）

通过包级变量 `var RT = NewRuntime()` 暴露单例。不使用 `sync.Once` 强制唯一性。

**理由**: 框架库应允许测试场景创建多个实例。Convention 单例足够表达"进程唯一"的语义，同时不限制测试灵活性。`main()` 中直接使用 `runtime.RT`。

**替代方案**: `sync.Once` / `init()` — 过于严格，无法在测试中重新创建，且无配置时机。已排除。

### D2: Logger 通过导出字段直接暴露

`FrameLogger` 和 `RpcLogger` 作为 `*logger.Logger` 类型导出字段，用户直接调用 `RT.FrameLogger.AddOutput(...)`。

**理由**: 最简洁。Logger 的 identify 在创建时固定（"framework" / "rpc"），Output 配置完全由用户控制，不需要额外的配置层。

**替代方案**: Options 结构体 / Configure 函数 / 闭包 — 增加不必要的抽象层，当前场景下直接暴露字段足够。

### D3: Component 注册表按 name，注册不启动

`RegisterComponent(name, comp)` 仅将 Component 存入 map，不调用 `Start()`。Worker 在 `Startup()` 中通过 `GetComponent(name)` 获取后调用 `ConnectComponent(ctx, comp)` 按需启动。

**理由**: 与 baseComponent 的 refCount 机制完美配合。多个 Worker 共享同一 Component 时，首次 Connect 启动，后续 Connect 只增加引用计数，最后一个 Stop 才真正断开。

### D4: Worker/Service 注册表按 id，卸载时 Stop

`InstallService(svc)` 以 `svc.GetMetadata().Id` 为 key 存入 map。`UninstallService(id)` 先调用 `svc.Stop()`，再从 map 中移除。

**理由**: Runtime 拥有注册表内实体的生命周期管理权。卸载 = 停止 + 移除，语义清晰，避免悬空资源。

### D5: NewRuntime 零参数构造

`NewRuntime()` 不接受任何参数。`startTime` 取 `time.Now()`，`root` 取 `os.Getwd()`，Logger 用固定 identify 创建。

**理由**: 所有数据都可内部推导，无需外部传入。保持 API 最简。

### D6: Startup/Shutdown 占位

`Startup(ctx context.Context) error` 和 `Shutdown() error` 方法体留空（直接返回 nil）。

**理由**: 为未来进程级生命周期编排预留接口（如按顺序启动所有 Service、优雅关停等），当前不实现逻辑。

## Risks / Trade-offs

- **[Convention 单例无强制保护]** → 用户可以创建多个 Runtime 实例。可接受，框架库不应限制测试场景。
- **[Component 注册表无生命周期管理]** → 注册的 Component 不会被 Runtime 自动启动或停止。这是设计意图（D3），Worker 负责。如果所有 Worker 都关停了但没人 unregister Component，它仍然只是注册状态（未启动），无资源泄漏。
- **[UninstallService 调用 Stop 可能失败]** → Stop 返回 error，UninstallService 应将该错误返回给调用者，但即使失败也从 map 中移除（避免残留）。需要决定是否在 Stop 失败时仍移除。当前倾向：失败也移除，记录日志。
- **[导出 *logger.Logger 字段]** → 用户可以替换 Logger 实例（赋值）。风险低，框架信任使用者。
