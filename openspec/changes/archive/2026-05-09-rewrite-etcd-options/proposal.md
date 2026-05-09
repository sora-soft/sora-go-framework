## Why

Go 端的 `EtcdOptions` 将 etcd 客户端配置平铺在顶层，缺少 `TTL`（Lease 租约时长）和 `Prefix`（Key 前缀命名空间）字段，且与 TypeScript 端 `IEtcdComponentOptions` 的嵌套结构不一致。这导致无法实现 Lease 管理、分布式锁、Key 路径构建等核心功能，也无法与 TS 端保持配置结构的对应关系。

## What Changes

- **BREAKING**: 将 `EtcdOptions` 重命名为 `EtcdComponentOptions`，采用嵌套结构：`Etcd`（客户端配置）、`TTL`、`Prefix`
- **BREAKING**: 提取独立的 `EtcdClientConfig` 类型，承载 etcd 连接配置
- **BREAKING**: `SetOptions` 中增加字段校验（`Endpoints` 非空、`DialTimeout > 0`、`TTL > 0`、`Prefix` 非空）
- `Connect()` 方法适配嵌套结构，从 `options.Etcd` 读取客户端配置
- 更新所有测试用例

## Capabilities

### New Capabilities

- `etcd-options`: Etcd 组件选项结构的定义与校验，包括嵌套的客户端配置、TTL、Prefix

### Modified Capabilities

（无现有 specs）

## Impact

- `pkg/component/etcd/etcd.go` — 选项类型重定义，`SetOptions` 校验逻辑，`Connect` 适配
- `pkg/component/etcd/etcd_test.go` — 所有测试用例更新
- 使用 `EtcdOptions` 的外部调用方需改为 `EtcdComponentOptions`
