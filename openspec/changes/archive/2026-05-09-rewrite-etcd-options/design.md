## Context

当前 Go 端 `EtcdOptions` 是一个平铺结构，直接包含 etcd 客户端连接字段（`Endpoints`、`DialTimeout` 等），缺少 `TTL` 和 `Prefix`。TypeScript 端的 `IEtcdComponentOptions` 采用嵌套结构：`etcd`（客户端配置）、`ttl`、`prefix`，三层关注点分离清晰。

Go 的 `clientv3.Config` 不可直接用于 JSON 序列化（含 `*tls.Config`、接口等字段），需要自定义可序列化的配置结构。

## Goals / Non-Goals

**Goals:**
- Go 端选项结构与 TS 端 `IEtcdComponentOptions` 保持一致的嵌套结构
- 独立导出 `EtcdClientConfig` 类型
- `SetOptions` 中对必填字段进行严格校验
- 保持与 `component.componentImpl` 接口的兼容

**Non-Goals:**
- 不实现 Lease 管理、分布式锁、Key 路径构建（这些依赖 TTL/Prefix，但属于后续变更）
- 不修改 `component.BaseComponent` 或 `component.componentImpl` 接口
- 不处理 TLS 配置（暂不纳入 EtcdClientConfig）

## Decisions

### 1. 嵌套结构而非平铺

**选择**: `EtcdComponentOptions` 包含嵌套的 `Etcd EtcdClientConfig` 字段。

**理由**: 与 TS 端一致；客户端配置与业务配置（TTL、Prefix）关注点分离；未来客户端配置可独立扩展（如加 TLS）而不污染业务字段。

**替代方案**: 平铺所有字段到顶层——Go 风格更常见，但字段增多后边界模糊，且与 TS 端不一致。

### 2. TTL 使用 int64（秒）

**选择**: `TTL int64` 表示秒数。

**理由**: JSON 配置文件中 `60` 比 `"1m0s"` 更直观；与 TS 端 `ttl: number`（秒）一致；`Connect` 中转换为 `time.Duration` 即可。

**替代方案**: `time.Duration`——序列化为纳秒，JSON 可读性差。

### 3. Prefix 必填，无默认值

**选择**: `SetOptions` 校验 `Prefix` 非空，空字符串时返回错误。

**理由**: 所有 Key 操作都依赖 prefix，缺少它会导致数据错乱；强制显式指定比隐式默认更安全。

### 4. SetOptions 集中校验

**选择**: 所有字段校验在 `SetOptions` 中完成，`Connect` 只负责使用已校验的配置。

**理由**: 尽早失败（fail-fast），避免运行时才发现配置错误。

## Risks / Trade-offs

- **BREAKING CHANGE**: `EtcdOptions` → `EtcdComponentOptions`，所有外部引用需更新 → 影响范围有限，etcd 组件尚未被广泛使用
- **嵌套 JSON 配置**: 配置文件需要 `{"etcd": {"endpoints": [...]}, "ttl": 60, "prefix": "/app"}` → 比平铺多一层，但层次清晰
