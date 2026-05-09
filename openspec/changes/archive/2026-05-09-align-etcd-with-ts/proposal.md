## Why

Go 版 `EtcdComponent` 和 `EtcdBackend` 与 TS 版 (`@sora-soft/etcd-component` + `@sora-soft/etcd-discovery`) 存在显著功能差距。当前 Go 版缺少 lease 全生命周期管理、自动重连、分布式锁、持久化写入等核心能力；Discovery 层缺少本地/远程数据分离、lease reconnect 后重注册、写操作序列化和 endpoint 关联校验。这导致 etcd 集群短暂不可用时，所有注册数据会静默丢失且永不恢复。

## What Changes

### EtcdComponent (`pkg/component/etcd/etcd.go`)
- **新增** Lease 全生命周期管理（grant / keepalive / revoke）
- **新增** Lease lost 检测与自动重连（exponential backoff, 100ms→30s）
- **新增** `OnLeaseReconnect` 回调注册机制
- **新增** `PersistPut` / `PersistDel` 持久化写入（重连后自动恢复）
- **新增** `Lock` 分布式锁（基于 `concurrency.NewMutex`）
- **新增** `Keys` 路径拼接辅助方法
- **修改** `Connect` 流程：创建 client 后 grant lease + 启动 keepalive
- **修改** `Disconnect` 流程：设置 destroyed 标志 → cancel keepalive → revoke lease → close client
- **修改** `BaseEtcdComponent` 代理所有新增方法

### EtcdBackend / Discovery (`pkg/discovery/store/etcd/`)
- **新增** `etcdRegistry` 本地 map 跟踪（localNodes / localServices / localWorkers / localEndpoints）
- **新增** lease reconnect 时遍历本地 map 重新注册所有实体
- **新增** `etcdRegistry` 写操作 mutex 序列化
- **新增** Endpoint 更新时校验关联 Service 是否存在
- **修改** `Election` 使用 `concurrency.Election` 替换手写简化实现
- **修改** Watcher 使用可取消 context 而非 `context.Background()`
- **修改** `EtcdBackend` 不再自行管理 lease，改为从 `EtcdComponent` 获取 leaseID
- **修改** `store.go` 清理 `println` 调试输出，替换为正式日志

### **BREAKING** Changes
- `EtcdComponent.Connect()` 行为变更：现在会自动 grant lease 并启动 keepalive
- `EtcdBackend.Connect()` 不再自行创建 lease，改为从 EtcdComponent 获取
- `BaseEtcdComponent` 新增方法（纯新增，不影响现有 API）

## Capabilities

### New Capabilities
- `etcd-lease-management`: EtcdComponent 自管理 lease 全生命周期（grant、keepalive、lost 检测、revoke），提供 reconnect 回调和持久化写入
- `etcd-distributed-lock`: 基于 etcd concurrency 包的分布式锁能力
- `discovery-reconnect-reregister`: EtcdBackend 在 lease reconnect 后自动重新注册所有本地实体

### Modified Capabilities
- `discovery-backend`: 需求变更——Backend 不再自行管理 lease，改为依赖 EtcdComponent 的 lease；Connect/Disconnect 流程相应调整
- `discovery-election`: 需求变更——Election 实现从手写简化版改为使用 etcd concurrency 包

## Impact

- **`pkg/component/etcd/`**: `etcd.go` 大幅重写，可能拆分为多个文件（lease.go, reconnect.go, lock.go, persist.go）
- **`pkg/discovery/store/etcd/`**: `backend.go`、`registry.go`、`election.go`、`store.go` 均需修改
- **`pkg/component/`**: `interface.go` 的 `componentImpl` 接口不变，无需修改
- **`pkg/discovery/`**: 顶层接口（Backend/Registry/Discovery/Election）不变，无需修改
- **依赖**: 新增 `go.etcd.io/etcd/client/v3/concurrency` 包引用
- **Runtime**: `runtime.go` 中 `InstallService` / `InstallWorker` 的调用方式不受影响
