## Context

`pkg/runner` 已有 `baseWorker`（管理 Worker 生命周期）和 `BaseService`（嵌入 baseWorker，空壳）。`pkg/component` 已建立了成熟的三层设计模式：公共接口 `Component` / hook 接口 `componentImpl` / 实现 `baseComponent`。当前 runner 缺少公共接口层、缺少组件/Listener 注册机制，与 component 模式不一致。

Runner 的外部实现者需要在 `Startup()` 中创建基础设施组件（etcd、MySQL）和 RPC Listener，并让框架接管这些资源的生命周期管理（连接/断开、启动/停止）。

## Goals / Non-Goals

**Goals:**
- 建立 `Worker` / `Service` 公共接口，与 `Component` 三层设计对称
- 提供 Ref 注入机制，让 Runner 在 Startup 中注册 Component 和 Listener
- Component 和 Listener 注册立即生效，框架在 Stop 时统一清理
- Startup 失败时自动回滚已注册的资源
- Service.Stop 先停 Listener 再走 Worker 关停流程

**Non-Goals:**
- 不实现具体的 Component（etcd、MySQL 等）
- 不实现 Runtime 单例的 Worker/Service 注册/创建逻辑
- 不实现 Worker/Service 间的依赖管理
- 不实现动态添加/移除 Component 或 Listener 的运行时管理（当前仅支持 Startup 中注册，但 ConnectComponent/InstallListener 方法本身可在任意时机调用）

## Decisions

### D1: 三层接口设计（与 Component 对称）

对外接口 `Worker` / `Service` 暴露给消费者，包含 Start/Stop/Go/GetMetadata。Hook 接口 `Runner` 由实现者提供（Startup/Shutdown）。`baseWorker` / `baseService` 持有 Runner 实例，实现对外接口。

**理由**: 与 `pkg/component` 模式完全一致，降低框架学习成本。消费者只需关心 Worker/Service 接口，实现者只需关心 Runner 接口。

### D2: Ref 注入模式（interface assertion）

通过 `WorkerRefAware` / `ServiceRefAware` 可选接口实现 Ref 注入。`NewWorker` / `NewService` 构造时检测 Runner 是否实现这些接口，若实现则注入对应 Ref。

**理由**: Go 标准库惯用模式（如 `io.Reader` + 可选 `io.WriterAt`）。不强制 Runner 实现 RefAware，但提供了即可获得框架能力。支持任意时机注册，不仅限于 Start 阶段。

**替代方案**: Provider 模式（Startup 后查询 Components()/Listeners()）——仅支持 Start 时注册，不支持动态注册，已排除。RunnerContext 模式（Startup 签名加参数）——破坏 Runner 接口兼容性，且 Service 难以复用 Worker 的 Start 流程，已排除。

### D3: ConnectComponent / InstallListener 立即生效

调用 `ConnectComponent` 时立即执行 `Component.Start(ctx)`，成功后记录到 components 切片。调用 `InstallListener` 时立即执行 `Listener.Start(ctx)`，成功后记录到 listeners 切片。失败的不记录。

**理由**: 简单直观，用户调用后即生效，无需等待后续统一处理。支持在 Startup 内或运行时任意时机调用。

### D4: Startup 失败自动回滚

当 `Runner.Startup()` 返回 error 时，`baseWorker.Start()` 自动调用 `disconnectComponents()` 断开所有已成功连接的 Component（仅断开成功记录的），然后设置 Error 状态。

**理由**: 避免资源泄漏。用户无需在 Startup 中手动处理组件回滚，框架保证要么全部成功进入 Ready，要么失败并清理已分配资源。

### D5: WorkerMetaData 统一大结构

`WorkerMetaData` 包含 Worker 和 Service 的全部字段（Labels、Listeners）。Worker 返回时这些字段为空（JSON omitempty 不输出），Service 返回时填充完整。

**理由**: Go 不支持协变返回类型，`Service` 嵌入 `Worker` 接口时 `GetMetadata()` 只能返回同一类型。统一结构比两个不同名称的方法更简洁。`omitempty` 确保 Worker 序列化时不会输出多余字段。

### D6: Service.Stop override Worker.Stop

`baseService.Stop()` 先调用 `stopListeners()` 停止所有 Listener，再走 Worker 的关停流程（cancel + wg.Wait + Runner.Shutdown + disconnectComponents）。

**理由**: Listener 是流量入口，应先停止接受新连接，再优雅关停内部业务和断开基础设施组件。

### D7: baseWorker / baseService 改为 unexported

struct 不导出，仅通过 `Worker` / `Service` 接口暴露给外部。

**理由**: 与 `baseComponent` 一致。隐藏实现细节，API 面向接口编程，未来可替换实现而不影响消费者。

## Risks / Trade-offs

- **[Runner 需手动实现 RefAware]** → 样板代码（一个字段 + 一个 setter 方法），但换来灵活性和框架管理能力。可接受。
- **[ConnectComponent 并发安全]** → components/listeners 切片需 mutex 保护，用户可能从不同 goroutine 动态注册。开销可忽略（注册频率低）。
- **[WorkerMetaData 有空字段]** → Worker 的 Labels/Listeners 为空值，但 omitempty 保证 JSON 输出干净。内存开销可忽略。
- **[BaseService.Stop 重复 baseWorker.Stop 逻辑]** → 无法通过 `s.baseWorker.Stop()` 调用后再补 stopListeners()，因为 stopListeners 必须在 cancel 之前执行。需要手动编排完整流程，存在与 baseWorker.Stop 的同步维护风险。
