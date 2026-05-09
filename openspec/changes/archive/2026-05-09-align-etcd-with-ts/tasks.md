## 1. EtcdComponent 文件拆分与结构体重构

- [x] 1.1 将 `pkg/component/etcd/etcd.go` 中的错误定义提取到 `errors.go`
- [x] 1.2 重构 `EtcdComponent` 结构体：新增 `lease clientv3.Lease`、`leaseID clientv3.LeaseID`、`keepAliveCtx context.Context`、`keepAliveCancel context.CancelFunc`、`persistValues map[string]string`、`onLeaseReconnect []LeaseReconnectFunc`、`reconnecting bool`、`destroyed bool`、`reconnectMu sync.Mutex` 字段
- [x] 1.3 新增 `LeaseReconnectFunc` 类型定义和 `OnLeaseReconnect` 方法
- [x] 1.4 新增 `LeaseID()`、`PersistPut()`、`PersistDel()`、`Keys()`、`Lock()` 方法签名到 `EtcdComponent`
- [x] 1.5 更新 `BaseEtcdComponent` 添加所有新增方法的代理

## 2. EtcdComponent Lease 管理

- [x] 2.1 创建 `lease.go`：实现 `grantLease(ctx)` 方法（`clientv3.NewLease` + `Grant`），在 `Connect()` 末尾调用
- [x] 2.2 实现 `startKeepAlive()` 方法：启动 keepalive goroutine，持续 drain keepalive channel
- [x] 2.3 实现 lease lost 检测：当 keepalive channel 关闭或返回 error response 时，触发 `reconnect()`
- [x] 2.4 修改 `Connect()` 流程：创建 client → 健康检查 → grantLease → startKeepAlive
- [x] 2.5 修改 `Disconnect()` 流程：设置 `destroyed=true` → cancel keepAliveCtx → revoke lease（忽略错误） → close client → 清空字段

## 3. EtcdComponent 自动重连

- [x] 3.1 创建 `reconnect.go`：实现 `reconnect(err)` 方法，使用 `reconnectMu` 串行化
- [x] 3.2 实现 exponential backoff 循环（100ms → 30s，无限重试），每轮：检查 destroyed → revoke 旧 lease → grantLease → startKeepAlive → 恢复 persistValues → 通知 onLeaseReconnect 回调
- [x] 3.3 实现 persistValues 恢复逻辑：遍历 persistValues map，用新 leaseID 重新写入每个 key-value

## 4. EtcdComponent 持久化与锁

- [x] 4.1 创建 `persist.go`：实现 `PersistPut(ctx, key, value)` — 写入 etcd（关联 leaseID）+ 存入 persistValues map
- [x] 4.2 实现 `PersistDel(ctx, key)` — 从 etcd 删除 + 从 persistValues map 移除
- [x] 4.3 创建 `lock.go`：实现 `Lock(ctx, key, fn, ttlSec)` — 使用 `concurrency.NewSession` + `concurrency.NewMutex` + `Lock/Unlock`

## 5. EtcdBackend Lease 依赖迁移

- [x] 5.1 修改 `EtcdBackend.Connect()`：移除自行创建 lease 的逻辑，改为从 `BaseEtcdComponent.LeaseID()` 获取 leaseID
- [x] 5.2 修改 `EtcdBackend.Disconnect()`：移除 lease revoke/close 逻辑，保留 watcher 关闭和 component Stop
- [x] 5.3 修改 `EtcdBackend.putWithLease()` 使用从 component 获取的 leaseID

## 6. EtcdRegistry 本地/远程数据分离

- [x] 6.1 为 `etcdRegistry` 添加 `localNodes`、`localServices`、`localWorkers`、`localEndpoints` 四个 map 和 `mu sync.Mutex`
- [x] 6.2 修改 `RegisterNode/Service/Worker/Endpoint`：加锁 → 写入 etcd → 存入对应 local map → 解锁
- [x] 6.3 修改 `UnregisterNode/Service/Worker/Endpoint`：加锁 → 删除 etcd key → 从 local map 移除 → 解锁
- [x] 6.4 实现 `reRegisterAll(ctx, leaseID)` 方法：遍历所有 local map，用新 leaseID 重新写入 etcd

## 7. EtcdBackend Reconnect 重注册

- [x] 7.1 在 `EtcdBackend.Connect()` 中注册 `OnLeaseReconnect` 回调
- [x] 7.2 回调逻辑：获取新 leaseID → 调用 `registry.reRegisterAll(ctx, newLeaseID)` → 逐个重注册，失败记录日志但不中断

## 8. Endpoint 关联校验

- [x] 8.1 修改 `store.updateEndpoint()`：在存入前检查 `s.services[meta.TargetID]` 是否存在，不存在则跳过

## 9. Watcher Context 管理

- [x] 9.1 为 `EtcdBackend` 添加 `watchCtx context.Context` 和 `watchCancel context.CancelFunc` 字段
- [x] 9.2 修改 `startWatchers()` 使用 `watchCtx` 替代 `context.Background()`
- [x] 9.3 修改 `Disconnect()` 调用 `watchCancel()` 以终止所有 watcher goroutine

## 10. Election 改造

- [x] 10.1 重写 `etcdElection` 结构体：使用 `concurrency.Session` 和 `concurrency.Election`，session 使用独立 lease
- [x] 10.2 重写 `Campaign(ctx, id)`：委托给 `concurrency.Election.Campaign(ctx, id)`
- [x] 10.3 重写 `Resign(ctx)`：委托给 `concurrency.Election.Resign(ctx)`
- [x] 10.4 重写 `Leader(ctx)`：委托给 `concurrency.Election.Leader(ctx)`，解析返回值
- [x] 10.5 重写 `Watch(ctx)`：使用 goroutine 定期轮询 leader 变化推送到 channel
- [x] 10.6 修改 `EtcdBackend.NewElection()` 构造新的 `etcdElection` 实例

## 11. 清理与日志

- [x] 11.1 清理 `store.go` 中所有 `println` 调试输出，替换为 `runtime.RT.FrameLogger` 调用
- [x] 11.2 在 `reconnect.go`、`lease.go` 中添加关键日志（lease lost、reconnect 开始/成功/失败、keepalive 错误）
