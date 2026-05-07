## 1. Provider 结构体变更

- [x] 1.1 在 Provider 结构体中添加 `refCount int` 字段

## 2. Start 方法改造

- [x] 2.1 将 Start 签名从 `Start(ctx context.Context)` 改为 `Start(ctx context.Context) error`
- [x] 2.2 在 Start 中实现引用计数逻辑：获取 mu 写锁，refCount > 0 时仅递增并返回 nil，否则执行原有启动逻辑并设置 refCount = 1

## 3. Stop 方法改造

- [x] 3.1 将 Stop 签名从 `Stop()` 改为 `Stop() error`
- [x] 3.2 在 Stop 中实现引用计数逻辑：获取 mu 写锁，refCount <= 0 时直接返回 nil，否则递减；refCount 降至 0 时执行原有清理逻辑

## 4. 调用方适配

- [x] 4.1 更新 `cmd/sora-test/main.go` 中 Provider.Start/Stop 的调用，适配新签名
- [x] 4.2 搜索并更新所有其他 Provider.Start/Stop 的调用点
