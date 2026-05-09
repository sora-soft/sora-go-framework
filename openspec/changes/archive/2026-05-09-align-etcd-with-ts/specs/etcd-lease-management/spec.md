## ADDED Requirements

### Requirement: Lease grant on connect
EtcdComponent 的 `Connect(ctx)` 方法 SHALL 在创建 etcd client 后立即 grant 一个 lease（TTL 由 options 指定），并启动 keepalive goroutine。leaseID SHALL 通过 `LeaseID()` 方法暴露给上层。

#### Scenario: Connect grants lease
- **WHEN** `Connect(ctx)` 成功执行
- **THEN** EtcdComponent SHALL 持有有效的 leaseID，且 `LeaseID()` SHALL 返回该 leaseID

#### Scenario: Connect without options set
- **WHEN** `Connect(ctx)` 在 `SetOptions` 之前调用
- **THEN** SHALL 返回 `ErrEtcdOptionsNotSet` 错误

### Requirement: Lease keepalive
EtcdComponent SHALL 在 lease grant 后启动一个后台 goroutine，持续发送 keepalive 请求以维持 lease 有效。keepalive goroutine SHALL 使用独立的 context，在 `Disconnect()` 时通过 cancel 该 context 来停止。

#### Scenario: Keepalive keeps lease alive
- **WHEN** EtcdComponent 已连接且 keepalive goroutine 正在运行
- **THEN** lease SHALL 持续保持有效状态，不会因 TTL 到期而过期

#### Scenario: Keepalive stops on disconnect
- **WHEN** `Disconnect()` 被调用
- **THEN** keepalive goroutine SHALL 停止

### Requirement: Lease lost detection
EtcdComponent SHALL 检测 lease lost 事件。当 keepalive response channel 关闭或返回错误时，SHALL 视为 lease lost 并触发自动重连流程。

#### Scenario: Keepalive channel closes
- **WHEN** keepalive response channel 关闭（etcd 集群不可用）
- **THEN** EtcdComponent SHALL 检测到 lease lost 并启动 reconnect 流程

#### Scenario: Keepalive returns error
- **WHEN** keepalive response 包含非空 Error 字段
- **THEN** EtcdComponent SHALL 检测到 lease lost 并启动 reconnect 流程

### Requirement: Auto-reconnect with exponential backoff
EtcdComponent SHALL 在 lease lost 后自动执行 reconnect 流程：revoke 旧 lease → grant 新 lease → 启动新 keepalive → 通知回调。重试 SHALL 使用 exponential backoff，初始间隔 100ms，最大间隔 30s，无限重试直到成功或组件被销毁。

#### Scenario: Reconnect succeeds on first try
- **WHEN** lease lost 后 etcd 集群立即可用
- **THEN** EtcdComponent SHALL 成功 grant 新 lease 并启动新 keepalive

#### Scenario: Reconnect retries with backoff
- **WHEN** 第一次 reconnect 失败
- **THEN** SHALL 在 100ms 后重试；第二次失败后在 200ms 后重试；依次翻倍，上限 30s

#### Scenario: Reconnect stops when destroyed
- **WHEN** `Disconnect()` 在 reconnect 循环中被调用
- **THEN** reconnect 循环 SHALL 立即终止，不再重试

### Requirement: LeaseReconnect callback
EtcdComponent SHALL 提供 `OnLeaseReconnect(fn func(leaseID clientv3.LeaseID, err error))` 方法，允许上层注册回调。reconnect 成功后 SHALL 按注册顺序调用所有回调，传入新 leaseID 和导致 reconnect 的原始错误。多并发的 reconnect SHALL 被串行化（同一时刻只有一个 reconnect 在执行）。

#### Scenario: Callback invoked on reconnect
- **WHEN** lease lost 后 reconnect 成功完成
- **THEN** 所有已注册的 `OnLeaseReconnect` 回调 SHALL 被调用，传入新 leaseID

#### Scenario: Multiple callbacks invoked in order
- **WHEN** 注册了 callback1 和 callback2
- **THEN** reconnect 后 callback1 SHALL 先于 callback2 被调用

#### Scenario: Concurrent reconnect prevention
- **WHEN** lease lost 触发 reconnect 时另一个 reconnect 正在进行
- **THEN** 第二个 reconnect SHALL 被忽略（不排队、不重复）

### Requirement: PersistPut
EtcdComponent SHALL 提供 `PersistPut(ctx, key, value) error` 方法。该方法 SHALL 将 key-value 写入当前 lease（关联 leaseID），同时将 key-value 存入内部 `persistValues` map。reconnect 后 SHALL 自动重新写入 `persistValues` 中的所有 key-value。

#### Scenario: PersistPut writes with lease
- **WHEN** `PersistPut(ctx, "service/abc", "data")` 被调用
- **THEN** key "service/abc" SHALL 写入 etcd 并关联当前 leaseID，同时存入 persistValues map

#### Scenario: PersistPut re-applied on reconnect
- **WHEN** reconnect 成功后
- **THEN** persistValues 中的所有 key-value SHALL 被重新写入新 lease

#### Scenario: PersistPut without lease
- **WHEN** `PersistPut` 在 lease 不存在时被调用
- **THEN** SHALL 返回 `ErrEtcdLeaseNotFound` 错误

### Requirement: PersistDel
EtcdComponent SHALL 提供 `PersistDel(ctx, key) error` 方法。该方法 SHALL 从 etcd 删除 key，并从 `persistValues` map 中移除。

#### Scenario: PersistDel removes key
- **WHEN** `PersistDel(ctx, "service/abc")` 被调用
- **THEN** key "service/abc" SHALL 从 etcd 删除，并从 persistValues map 中移除

### Requirement: Keys helper
EtcdComponent SHALL 提供 `Keys(args ...string) string` 方法，返回 options.Prefix 与所有 args 用 `/` 拼接的路径。

#### Scenario: Keys joins paths
- **WHEN** `Keys("service", "abc")` 被调用且 prefix 为 "sora"
- **THEN** SHALL 返回 "sora/service/abc"

### Requirement: Disconnect cleans up
EtcdComponent 的 `Disconnect()` SHALL 依次执行：设置 destroyed 标志 → cancel keepalive context → revoke lease（忽略错误） → close client。disconnect 后 reconnect 循环 SHALL 终止。

#### Scenario: Disconnect stops reconnect
- **WHEN** reconnect 正在进行时 `Disconnect()` 被调用
- **THEN** reconnect 循环 SHALL 检测到 destroyed 标志并终止

#### Scenario: Disconnect revokes lease
- **WHEN** `Disconnect()` 被调用且 lease 有效
- **THEN** lease SHALL 被 revoke（错误被忽略，因为 lease 可能已 lost）
