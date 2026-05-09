## Why

EtcdBackend.Connect() 从 runtime 获取组件后直接访问 Client()，但从未调用 component.Start() 确保组件已连接。如果调用者没有提前启动组件，Client() 返回 nil，导致一个不明确的 "not connected" 错误。同时 Disconnect() 也没有调用 component.Stop() 释放引用计数。此外 initFromEtcd 中残留了 debug println 语句。

## What Changes

- EtcdBackend.Connect() 在获取组件后增加调用 component.Start(ctx)，利用 baseComponent 的引用计数机制确保组件已连接
- EtcdBackend.Disconnect() 增加调用 component.Stop()，正确减少引用计数
- EtcdBackend 结构体新增字段保存 Component 引用
- 清除 initFromEtcd 中的 debug println 语句

## Capabilities

### New Capabilities

_(无新增功能)_

### Modified Capabilities

- `discovery-backend`: 补充 Backend 连接生命周期要求 — Connect SHALL 确保底层组件已启动，Disconnect SHALL 释放组件引用

## Impact

- `pkg/discovery/store/etcd/backend.go` — 主要修改文件
- `pkg/discovery/store/etcd/etcd_test.go` — 现有测试需要调整（setupTestComponent 中的手动 Start 可移除，由 Backend 自管理）
- `openspec/specs/discovery-backend/spec.md` — 增量规范
