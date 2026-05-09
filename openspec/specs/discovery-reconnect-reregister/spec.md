## ADDED Requirements

### Requirement: Local entity tracking
etcdRegistry SHALL 维护四个本地 map：`localNodes`、`localServices`、`localWorkers`、`localEndpoints`（key 为 entity ID，value 为完整的 Meta 结构体）。`Register*` 方法 SHALL 在写入 etcd 成功后将 Meta 存入对应的 local map。`Unregister*` 方法 SHALL 从 local map 中移除对应条目。

#### Scenario: Register stores in local map
- **WHEN** `RegisterService(ctx, serviceMeta)` 成功写入 etcd
- **THEN** serviceMeta SHALL 同时存入 `localServices[serviceMeta.ID]`

#### Scenario: Unregister removes from local map
- **WHEN** `UnregisterService(ctx, id)` 成功执行
- **THEN** `localServices[id]` SHALL 被删除

### Requirement: Reconnect re-registration
EtcdBackend SHALL 注册 `OnLeaseReconnect` 回调。回调被触发时，SHALL 遍历 etcdRegistry 的所有 local map（localNodes → localServices → localWorkers → localEndpoints），对每个实体调用 `putWithLease` 重新写入 etcd。任何实体的重注册失败 SHALL 记录错误日志但不中断后续实体的重注册。

#### Scenario: All local entities re-registered
- **WHEN** lease reconnect 回调被触发且 localServices 包含 service-A 和 service-B
- **THEN** service-A 和 service-B SHALL 都被重新写入 etcd（关联新 leaseID）

#### Scenario: Partial failure continues
- **WHEN** 重注册 service-A 失败
- **THEN** SHALL 记录错误日志，然后继续重注册 service-B

### Requirement: Write serialization
etcdRegistry 的所有 register/unregister 方法 SHALL 在 `sync.Mutex` 保护下执行，确保同一时刻只有一个写操作在执行。

#### Scenario: Concurrent registers serialized
- **WHEN** 两个 goroutine 同时调用 `RegisterService`
- **THEN** 两个操作 SHALL 串行执行，不会并发写入 etcd
