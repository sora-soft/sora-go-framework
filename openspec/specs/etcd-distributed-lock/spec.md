## ADDED Requirements

### Requirement: Distributed lock
EtcdComponent SHALL 提供 `Lock(ctx context.Context, key string, fn func() error, ttlSec int) error` 方法。该方法 SHALL 使用 `go.etcd.io/etcd/client/v3/concurrency` 包创建 session 和 mutex，获取锁后执行 fn，fn 执行完毕后释放锁。key SHALL 与 options.Prefix 拼接形成完整的锁路径。

#### Scenario: Lock acquires and executes
- **WHEN** `Lock(ctx, "job-1", fn, 5)` 被调用且无其他持有者
- **THEN** SHALL 获取名为 "<prefix>/job-1" 的锁，执行 fn，然后释放锁

#### Scenario: Lock blocks when held by another
- **WHEN** `Lock(ctx, "job-1", fn, 5)` 被调用且另一个进程持有同名锁
- **THEN** SHALL 阻塞直到锁可用，然后执行 fn

#### Scenario: Lock without client
- **WHEN** `Lock` 在 `Connect` 之前被调用
- **THEN** SHALL 返回 `ErrEtcdNotConnected` 错误

#### Scenario: Lock with context cancellation
- **WHEN** ctx 被取消时锁仍在等待
- **THEN** SHALL 返回 context 错误，不执行 fn
