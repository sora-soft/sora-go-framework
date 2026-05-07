## Context

当前 `pkg/rpc/provider` 包中 `Provider` 是一个导出的具体 struct。Worker 和 Service 层有 `ConnectComponent` 方法来管理 Component 的生命周期，但没有对应的方法来管理 Provider。用户需要自行处理 Provider 的启动和停止。

现有模式参考：
- `component.Component` interface + `baseComponent`（不导出）通过 `ConnectComponent` 注册到 `baseWorker`
- `rpc.Listener`（具体类型）通过 `InstallListener` 注册到 `baseService`
- Component 和 Provider 都使用引用计数（refCount）来支持多消费者共享同一实例

## Goals / Non-Goals

**Goals:**
- 将 Provider 重构为 interface，使 RPC 调用能力可替换、可 mock
- 在 runner 层提供统一的 Provider 生命周期管理（RegisterProvider + 自动停止）
- 支持 Provider 实例在多个 Worker/Service 间共享（通过已有 refCount）
- 与 Component 的管理模式保持一致

**Non-Goals:**
- 不修改 Provider 的 RPC 调用逻辑（selectSender、watchLoop 等）
- 不修改 Component 相关代码
- 不引入通用的 Lifecycle interface 来统一 Component 和 Provider

## Decisions

### Decision 1: Provider interface 包含 CallRpc

**选择**: Provider interface 包含 Start、Stop、CallRpc 三个方法，放在 `pkg/rpc/provider` 包中

**备选方案**:
- A: 只放 Start/Stop（最小接口），CallRpc 留在具体类型上
- B: 拆分为 Provider（生命周期）+ RpcProvider（含 CallRpc）两层 interface
- C: 全量放在一个 interface 上

**理由**: 跟 Component 模式对齐——Component interface 包含所有业务方法（Start/Stop/LoadOptions/GetMetaInfo），用户代码只需持有 interface。Provider interface 包含 CallRpc 使业务代码无需依赖具体类型，同时支持 mock 和替换。所有相关类型（CallOption、packet.Packet）已在同一个包内，不引入跨包依赖。

### Decision 2: Concrete 命名为 rpcProvider（不导出）

**选择**: 具体实现重命名为 `rpcProvider`，不导出。`NewProvider()` 返回 `Provider` interface。

**理由**: 与 Component 模式一致（Component interface + baseComponent 不导出）。防止外部代码依赖具体实现细节。

### Decision 3: RegisterProvider 放在 baseWorker 层

**选择**: `RegisterProvider` 方法放在 `baseWorker` 上，`baseService` 通过嵌入自动继承。

**理由**: Worker 和 Service 都可能需要通过 Provider 调用远程 RPC。Service 是 Worker 的扩展，不需要在 Service 层单独实现。这与 `ConnectComponent` 放在 baseWorker 的位置一致。

### Decision 4: Stop 顺序

**选择**: 
- Worker: `cancel → wg.Wait → Shutdown → stopProviders → disconnectComponents → Stopped`
- Service: `stopListeners → cancel → wg.Wait → Shutdown → stopProviders → disconnectComponents → Stopped`

**理由**: Provider 关闭意味着失去 RPC 调用能力。放在 Shutdown 之后确保 Shutdown 过程中仍可调用 RPC。放在 disconnectComponents 之前，因为 Components 可能依赖 RPC 能力。

### Decision 5: 删除 RpcSender 中未使用的 provider 字段

**选择**: 从 RpcSender struct 中删除 `provider *Provider` 字段，以及 NewRpcSender 中对应的参数。

**理由**: 经代码审查确认 RpcSender 从未使用该字段。

## Risks / Trade-offs

- **BREAKING**: `provider.NewProvider()` 返回类型从 `*Provider` 变为 `Provider` interface。使用方如果做了类型断言到 `*Provider` 会编译失败 → 缓解：正常使用（调用方法）不受影响
- **BREAKING**: `NewRpcSender` 签名变更，删除 provider 参数 → 缓解：RpcSender 是包内部使用，不影响外部
- **RpcSender 的 provider 字段删除后如需回退** → 缓解：可以重新添加，影响范围小
