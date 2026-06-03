## Context

当前 `rpcProvider.watchLoop()` 通过 `disco.WatchEndpoints()` 接收 endpoint 快照。对每个 endpoint，仅根据 ID 判断"新增"或"已存在"两种状态——新增则创建 `RpcSender`，已存在则跳过。不存在第三种分支："ID 相同但地址变了"。

这导致：当 listener 初始注册了一个不完整地址（如 `192.168.10.158:0`），后续 listener Ready 后更新为正确地址（如 `192.168.10.158:19347`），Provider 不会重建 sender，已建立的 `RpcSender` 的 `endpoint.Endpoint` 字段始终保持旧值。

## Goals / Non-Goals

**Goals:**
- Provider 能检测已有 endpoint 的地址/协议/codec 变更
- 检测到变更后，热替换 RpcSender（销毁旧的，创建新的）
- 保证替换过程中线程安全

**Non-Goals:**
- 不修改 LifeCycle 通知机制（channel 丢弃问题是独立议题）
- 不修改 Discovery Store 的推送机制
- 不处理 Weight/Labels 等非连接相关字段的变更（这些不影响连接）

## Decisions

### 决策 1：变更检测粒度 — 比较关键字段

比较 `Endpoint`、`Protocol`、`Codecs` 三个字段。只有这些字段变化才需要重建 sender（它们决定了连接目标、传输协议和编解码方式）。

**备选方案：**
- 深度比较整个 EndpointMeta struct → 会因 Weight/Labels 变化触发不必要的重建
- 比较 Endpoint 字段串 → 简单但不够，Protocol 或 Codecs 变化也需要重建

### 决策 2：热替换策略 — destroy + create

检测到变更时，先销毁旧 sender（cancel context、关闭连接、fail pending），再创建新 sender。

**备选方案：**
- 就地更新 sender 的 endpoint 字段 → 需要 sender 支持"重连到新地址"，增加 sender 的复杂度
- destroy + create → 复用现有的 `removeSenderLocked` + `addSenderLocked`，改动最小

### 决策 3：比较函数的位置 — 放在 provider 包内

新增一个包级函数 `endpointChanged(existing discovery.EndpointMeta, new discovery.EndpointMeta) bool`，逻辑清晰且易于测试。

## Risks / Trade-offs

- **[短暂连接中断]** → 热替换期间正在进行的 RPC 调用会收到 `ErrConnectionLost`，调用方需重试。这是可接受的，因为地址变更本身就是连接性质的变更。
- **[高频变更抖动]** → 如果外部频繁更新 endpoint 地址，可能导致 sender 反复重建。缓解：变更只比较关键字段，Weight/Labels 等变化不触发重建。
