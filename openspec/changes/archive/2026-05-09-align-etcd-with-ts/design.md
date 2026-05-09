## Context

Go 版 sora-framework 的 etcd 相关实现（`EtcdComponent` + `EtcdBackend`）与 TS 版存在结构性差距。当前状态：

- `EtcdComponent` 仅包装了 `clientv3.Client` 的创建和关闭，不管理 lease
- `EtcdBackend` 在 `Connect()` 中自行创建 lease + keepalive，但无 reconnect 逻辑
- lease 过期后所有注册数据静默丢失，永不恢复
- Election 是手写的简化版，不使用 etcd 的 revision-based 竞争机制
- Registry 不区分本地/远程数据

TS 版的架构是：`EtcdComponent` 自包含管理 lease 全生命周期 + 重连 + 事件通知，`ETCDDiscovery` 监听 `LeaseReconnect` 事件后重新注册本地数据。

约束：
- Go 版 `component.Component` 接口和 `discovery.Backend/Registry/Discovery/Election` 接口不变
- `BaseComponent` 的 refCount 机制不变
- `runtime.Runtime` 的调用方式不变
- 新增 `go.etcd.io/etcd/client/v3/concurrency` 依赖

## Goals / Non-Goals

**Goals:**
- EtcdComponent 自管理 lease 全生命周期（grant → keepalive → lost detection → reconnect → re-grant）
- EtcdComponent 提供回调机制通知上层 lease reconnect 事件
- EtcdComponent 提供持久化写入能力（reconnect 后自动恢复）
- EtcdComponent 提供分布式锁
- EtcdBackend 在 lease reconnect 后自动重新注册所有本地实体
- EtcdBackend 的 Registry 区分本地（我注册的）和远程（别人注册的）数据
- Registry 写操作序列化
- Endpoint 更新时校验关联 Service 存在
- Election 使用 etcd concurrency 包
- Watcher 使用可取消 context

**Non-Goals:**
- 不修改 `component.Component` / `componentImpl` 接口
- 不修改 `discovery.Backend/Registry/Discovery/Election` 顶层接口
- 不修改 `runtime.Runtime` 的调用方式
- 不实现 TS 版的 `QueueExecutor`（Go 版用 `sync.Mutex` 替代）
- 不实现 TS 版的 `RxJS Subject`（Go 版已有 `notifier` channel 模式）
- 不实现 TS 版的 `SubscriptionManager`（Go 版用 context cancellation 管理）
- 不引入额外的第三方依赖（除 etcd concurrency 包外）

## Decisions

### D1: Lease 管理归属 EtcdComponent

**决策**: 将 lease 管理从 `EtcdBackend` 移入 `EtcdComponent`，对齐 TS 版架构。

**理由**: TS 版 `EtcdComponent` 是 lease 的 owner。它 grant lease、监听 lost、自动重连、通知上层。`ETCDDiscovery` 只是"借用"这个 lease 注册数据。这种分层让 EtcdComponent 成为自包含的基础设施组件，可被多种上层消费者复用。

**备选方案**:
- B: lease 留在 EtcdBackend，在 Backend 层补齐 reconnect → 不符合 TS 架构，且 Backend 和 Component 的职责会纠缠不清

**影响**: `EtcdBackend.Connect()` 不再自行 `clientv3.NewLease` + `Grant`，改为从 `EtcdComponent` 获取 `LeaseID`。`EtcdBackend.Disconnect()` 不再自行 `lease.Close()`。

### D2: Lease lost 检测方式

**决策**: 通过 keepalive response channel 检测 lease lost。当 channel 关闭或返回 error response 时，触发 reconnect。

**理由**: Go 版 `clientv3` 的 lease 没有 TS 版 `etcd3` 的 `'lost'` 事件。最接近的机制是 keepalive channel：
```go
ch, _ := lease.KeepAlive(ctx, leaseID)
for resp := range ch {
    if resp.Error != nil { /* lease lost */ }
}
// channel closed → lease lost
```

**备选方案**:
- B: 定期 `TimeToLive` 轮询 → 增加额外网络开销，不够及时

### D3: Reconnect 回调机制

**决策**: 使用回调函数切片，而非通用 EventEmitter。

**理由**: 当前只有 `LeaseReconnect` 一个事件。Go 惯用法是回调函数（如 `http.HandlerFunc`、`database/sql` 的 `SetMaxOpenConns` 等）。引入通用 EventEmitter 会过度设计。

```go
type LeaseReconnectFunc func(leaseID clientv3.LeaseID, err error)

type EtcdComponent struct {
    // ...
    onLeaseReconnect []LeaseReconnectFunc
}

func (e *EtcdComponent) OnLeaseReconnect(fn LeaseReconnectFunc) {
    e.onLeaseReconnect = append(e.onLeaseReconnect, fn)
}
```

### D4: 重连退避策略

**决策**: 手写 exponential backoff（100ms → 30s），不引入额外依赖。

```go
interval := 100 * time.Millisecond
for {
    if e.destroyed { return }
    err := e.tryReconnect()
    if err == nil { return }
    log.Warn("reconnect failed, retrying in %v", interval)
    time.Sleep(interval)
    interval = min(interval*2, 30*time.Second)
}
```

**理由**: 逻辑简单（一个 for 循环），不值得引入 retry 库。TS 版的 `Retry` 工具也是框架内部的。

### D5: 本地/远程数据分离

**决策**: `etcdRegistry` 维护 `localNodes/localServices/localWorkers/localEndpoints` map（key 为 entity ID），存储完整的 Meta 数据。reconnect 时遍历这些 map 重新注册。远程数据继续由 `store` 管理（watcher 更新，discovery 查询）。

**理由**: 对齐 TS 版的 `localServiceIdMap_` / `remoteServiceIdMap_` 分离模式。

```
┌─ etcdRegistry ─────────────────┐   ┌─ store ─────────────────────┐
│ localNodes    map[string]Meta   │   │ nodes    map[string]Meta    │
│ localServices map[string]Meta   │   │ services map[string]Meta    │
│ localWorkers  map[string]Meta   │   │ workers  map[string]Meta    │
│ localEndpoints map[string]Meta  │   │ endpoints map[string]Meta   │
│ mu sync.Mutex                   │   │ mu sync.RWMutex             │
└─────────────────────────────────┘   └─────────────────────────────┘
     ↑ register/unregister 写入           ↑ watcher 更新 / discovery 读取
     ↑ reconnect 时遍历重注册
```

### D6: 写操作序列化

**决策**: `etcdRegistry` 使用 `sync.Mutex` 序列化所有 register/unregister 操作。

**理由**: TS 版用 `QueueExecutor`（单消费者队列）序列化。Go 版用 Mutex 更轻量，且 register/unregister 本身是短操作（一次 etcd put/delete），Mutex 足够保证一致性。

### D7: Election 使用 concurrency 包

**决策**: 使用 `go.etcd.io/etcd/client/v3/concurrency` 的 `Election` 替换手写实现。

**理由**: 手写 `etcdElection` 没有使用 revision-based 竞争，在并发和网络分区场景下不安全。`concurrency.Election` 是 etcd 官方提供的选举实现，基于 MVCC revision，有正确的 campaign/proclaim/resign 语义。

**注意**: 当前 `discovery.Election` 接口的 `Campaign(ctx, id)` 将 id 作为 val 传入，与 `concurrency.Election.Campaign(ctx, val)` 语义一致。`Resign()` 也一致。`Leader()` 需要从 `election.Leader(ctx)` 获取。`Watch()` 需要通过 observer 模式实现——定期查询或使用 `concurrency.NewWatcher`。

### D8: EtcdComponent 文件拆分

**决策**: 将 `etcd.go` 拆分为：
```
pkg/component/etcd/
├── etcd.go         — 结构体定义、options、Connect/Disconnect、公共 getter
├── lease.go        — grantLease、keepAlive goroutine、lease lost 检测
├── reconnect.go    — reconnect 循环、backoff、回调通知
├── lock.go         — 分布式锁
├── persist.go      — persistPut/persistDel
└── errors.go       — 错误定义
```

**理由**: 单文件 etcd.go 已有 166 行，加上新功能会超过 500 行。按职责拆分符合 Go 惯例（如 `net/http` 的 `client.go` / `server.go` / `transport.go`）。

## Risks / Trade-offs

[R1: reconnect 期间的短暂不可用] → 在 reconnect 循环中所有写入操作会失败（因为 lease 无效）。上层应自行处理临时错误。reconnect 成功后会自动恢复持久化数据和重新注册本地实体。

[R2: concurrency.Election 的 Session 依赖 lease] → `concurrency.NewSession` 创建的 session 也依赖 lease。如果 EtcdComponent 的 lease lost，election session 也会失效。需要在 reconnect 时重建 election session。或者 election 使用独立的 session（独立 lease），这样更稳健但增加 lease 数量。→ 建议 election 使用独立 session。

[R3: BaseEtcdComponent 类型断言] → `EtcdBackend` 当前通过 `c.(*etcdcomp.BaseEtcdComponent)` 类型断言获取 client。新增方法后，Backend 可能需要通过 Impl() 访问更多能力。→ 在 `BaseEtcdComponent` 上暴露所有需要的代理方法。

[R4: Watcher 在 reconnect 期间可能丢失事件] → etcd client 本身会处理 watch 重连（基于 revision），但 EtcdComponent 的 client close/reopen 期间可能丢失。→ TS 版也没有完美解决这个问题，目前可接受。未来可考虑使用 etcd client 的内置 watch recovery。

[R5: store.go 中 println 调试输出] → 替换为 `runtime.RT.FrameLogger`，引入 `pkg/runtime` 依赖。
