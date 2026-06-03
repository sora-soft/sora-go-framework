### 需求:Endpoint 变更检测
Provider 的 watchLoop 在处理 discovery 快照时，对于已存在的 endpoint ID，必须比较其 Endpoint、Protocol、Codecs 三个关键字段是否发生变化。当任一字段不同时，必须判定为"需要替换"。

#### 场景:Endpoint 地址变更
- **当** discovery 快照中某 endpoint ID 已存在于 senders 中，且 Endpoint 字段与当前 sender 的 endpoint.Endpoint 不同
- **那么** Provider 必须销毁旧 sender 并用新 endpoint 数据创建替换 sender

#### 场景:Endpoint 协议变更
- **当** discovery 快照中某 endpoint ID 已存在于 senders 中，且 Protocol 字段与当前 sender 的 endpoint.Protocol 不同
- **那么** Provider 必须销毁旧 sender 并用新 endpoint 数据创建替换 sender

#### 场景:Endpoint 编解码列表变更
- **当** discovery 快照中某 endpoint ID 已存在于 senders 中，且 Codecs 字段与当前 sender 的 endpoint.Codecs 不同
- **那么** Provider 必须销毁旧 sender 并用新 endpoint 数据创建替换 sender

#### 场景:Endpoint 非关键字段变更
- **当** discovery 快照中某 endpoint ID 已存在于 senders 中，仅有 Weight、Labels、State 等非连接相关字段变更
- **那么** Provider 禁止重建 sender，保持现有 sender 不变

#### 场景:Endpoint 无变更
- **当** discovery 快照中某 endpoint ID 已存在于 senders 中，且所有关键字段均相同
- **那么** Provider 禁止重建 sender，保持现有 sender 不变

### 需求:Sender 热替换
当检测到 endpoint 变更时，Provider 必须先销毁旧 RpcSender，再创建新 RpcSender，此过程必须在互斥锁保护下完成。

#### 场景:热替换顺序
- **当** endpoint 变更被检测到
- **那么** Provider 必须先调用旧 sender 的 Destroy，再创建新 sender 并 Start，全程持有 p.mu 锁

#### 场景:热替换期间进行中的 RPC
- **当** sender 被热替换，且旧 sender 上有未完成的 RPC 调用
- **那么** 旧 sender 的 Destroy 必须通过 failAllPending 通知所有等待中的调用方（返回 ErrConnectionLost）

### 需求:变更比较函数
Provider 包必须提供 `endpointChanged` 函数用于比较两个 EndpointMeta 的关键字段是否不同。

#### 场景:Endpoint 字符串不同
- **当** a.Endpoint != b.Endpoint
- **那么** endpointChanged 返回 true

#### 场景:Protocol 不同
- **当** a.Protocol != b.Protocol
- **那么** endpointChanged 返回 true

#### 场景:Codecs 切片不同
- **当** a.Codecs 与 b.Codecs 的长度或内容不同
- **那么** endpointChanged 返回 true

#### 场景:所有关键字段相同
- **当** a 和 b 的 Endpoint、Protocol、Codecs 完全一致
- **那么** endpointChanged 返回 false
