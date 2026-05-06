## Context

sora-go-framework 的 RPC 层当前只有服务端（Listener → Connection → OnRequest/OnNotify 回调），缺少客户端 RPC 调用能力。现有 Packet 体系将 `json.RawMessage` 作为 Payload 类型，导致：1) 用户必须手动序列化/反序列化；2) 编解码格式与 JSON 强耦合，无法适配 Protobuf 等。

本次变更同时解决两个问题：重新设计 Packet/Codec 体系实现编解码无关，以及构建三层 RPC 发送端框架（Provider → RpcSender → Connection）。

## Goals / Non-Goals

**Goals:**
- 统一 Packet 类型，通过泛型 `Decode[T]()` 隐藏 `[]byte` payload 细节，用户 API 完全类型安全
- 定义 `packet.PayloadCodec` 接口，Codec 体系编解码无关（JSON/Protobuf/...）
- 构建 RpcSender：管理单个 endpoint 的 Connection，指数退避重连（初始 500ms，上限 10s），pending request map + request ID 关联
- 构建 Provider：通过 `WatchEndpoints` 监听 discovery，管理 RpcSender 生命周期，权重随机选择，泛型 `CallRpc[Resp any]()` 方法
- 支持 `WithTarget(serviceId)` 指定调用特定服务实例
- 保持现有 JSON 线路格式不变（向后兼容）

**Non-Goals:**
- 服务端 handler 泛型注册（`listener.Handle[Req, Resp]`）——后续变更处理
- Transport 注册表/工厂——当前只有 TCP，硬编码即可
- Worker.ConnectProvider 集成——后续变更处理
- 连接池/多路复用——当前每 endpoint 一个 Connection
- 重试/熔断/限流策略——由调用方决定

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Worker / 应用层                                             │
│    resp, err := provider.CallRpc[MyResp](ctx, "method", req) │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  Provider                                                   │
│    serviceName: string                                      │
│    senders:      map[endpointId]*RpcSender                   │
│    sendersBySvc: map[serviceId][]*RpcSender                  │
│    disco.WatchEndpoints → 创建/销毁 RpcSender                │
│    CallRpc: 权重随机 或 WithTarget → RpcSender.callRpcRaw    │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  RpcSender                                                  │
│    endpoint: EndpointMeta                                   │
│    conn: *Connection (当前活跃连接)                           │
│    pending: map[requestId]chan Packet                        │
│    connectLoop: 创建 Connection → 监听状态 → 断开重连         │
│    handleResponse: requestId 查找 → 发送到 pending channel   │
│    callRpcRaw: 生成 requestId → 注册 pending → Send → 等待   │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  Connection (已有)                                           │
│    OnResponse: 新增回调，分发 Response Packet                │
│    SendRaw: 使用 Codec.EncodePacket 编码 Packet              │
└─────────────────────────────────────────────────────────────┘
```

## Decisions

### D1: Packet 统一为单一类型 + 泛型 Decode

**选择**: 删除四个独立 packet 类型，统一为一个 `Packet` struct，Payload 使用私有 `[]byte`，对外暴露 `Decode[T any]()` 泛型方法。

```go
// packet 包
type PayloadCodec interface {
    Marshal(v any) ([]byte, error)
    Unmarshal(data []byte, v any) error
}

type Packet struct {
    Opcode  PacketOpcode
    Method  string
    Service string
    Headers map[string]string
    payload []byte       // 私有，用户不可见
    codec   PayloadCodec // 私有，用于 Decode
}

func (p Packet) Decode[T any]() (T, error) { ... }
```

**备选方案**:
- A) `Payload any`（Connection 解码到 map[string]interface{}）—— JSON 可行但 Protobuf 不行，且类型断言不安全
- B) 泛型 Packet[T]——传染 Connection/Listener/Codec 全部泛型化，改动不可控

**理由**: `[]byte` 是最通用的原始类型，适配所有编解码格式。私有化 + `Decode[T]()` 让用户永远不接触原始字节。`PayloadCodec` 接口放在 `packet` 包避免循环 import。

### D2: Codec 接口重新设计

**选择**: 新接口包含 `Marshal/Unmarshal`（payload 级别）+ `EncodePacket/DecodePacket`（信封级别）。

```go
type Codec interface {
    GetCode() string
    Marshal(v any) ([]byte, error)
    Unmarshal(data []byte, v any) error
    EncodePacket(pkt packet.Packet) ([]byte, error)
    DecodePacket(data []byte) (packet.Packet, error)
}
```

**备选方案**:
- A) 保持旧 Encode/Decode + 新增 Marshal/Unmarshal——两套 API 共存增加认知负担

**理由**: 职责清晰：`Marshal/Unmarshal` 处理单个值的编解码，`EncodePacket/DecodePacket` 处理完整线路信封。JSON Codec 内部使用 wire struct 维持现有线路格式。

### D3: Connection 回调统一为 Packet

**选择**: `OnRequest`/`OnNotify`/`OnResponse` 全部使用 `func(conn *Connection, pkt Packet)`。

**理由**: 统一类型简化 Connection 内部实现（handlePacket 直接分发），消费方通过 `pkt.Opcode` 和 `pkt.Decode[T]()` 获取类型化数据。Command packet 内部处理（ping/pong），不暴露给用户回调。

### D4: RpcSender 无独立 LifeCycle

**选择**: RpcSender 不设 LifeCycle 状态机，完全跟随底层 Connection 状态。`isReady()` 直接检查 `conn.LifeCycle.GetState() == ConnectorStateReady`。

**理由**: RpcSender 的状态就是 Connection 的状态，额外状态机增加复杂度但不增加价值。connectLoop 管理连接生命周期即可。

### D5: 重连策略——指数退避 + 无限重试

**选择**: 初始延迟 500ms，每次翻倍，上限 10s，无限重试直到 endpoint 从 discovery 消失。

**理由**: 服务端可能临时不可用，指数退避避免风暴。无限重试保证最终一致。endpoint 消失是唯一的终止条件。

### D6: callRpc 同步化通过 pending map + channel

**选择**: `callRpcRaw` 生成 requestId，注册 `chan Packet`（容量 1），发送请求，select 等待 channel 或 context 取消。

```go
requestId := generateRequestId()
ch := make(chan Packet, 1)
pending[requestId] = ch
conn.SendRaw(ctx, pkt)
select {
case res := <-ch: return res, nil
case <-ctx.Done(): delete pending; return ErrTimeout
case <-senderCtx.Done(): delete pending; return ErrStopped
}
```

**理由**: channel 容量 1 保证 handleResponse 非阻塞写入。context 取消时清理 pending 防泄漏。

### D7: Provider.watchEndpoints 过滤策略

**选择**: 使用 `discovery.WatchEndpoints()` 获取全量快照，本地按 `TargetName == serviceName` 过滤。

**备选方案**:
- A) 新增 `WatchEndpointsByTargetName()`——需要修改 discovery 接口，超出本次范围

**理由**: WatchEndpoints 已有全量推送，本地过滤开销极小。未来可按需添加专用 watch 方法。

### D8: Request ID 使用 crypto/rand 生成

**选择**: `crypto/rand.Read` 生成 16 字节随机数，hex 编码为 32 字符字符串，存入 `headers["x-sora-rpc-id"]`。

**理由**: UUID 库引入额外依赖，crypto/rand 足够唯一且无外部依赖。

## Packet 线路格式

JSON 线路格式保持不变（向后兼容），Codec 内部通过 wire struct 映射：

| Packet 类型 | Method 字段映射 | Payload 字段映射 |
|------------|----------------|-----------------|
| Request    | `method`       | `payload`       |
| Response   | _(空)_         | `payload`       |
| Notify     | `method`       | `payload`       |
| Command    | `command`      | `args`          |

## Data Flow

### CallRpc 完整流程

```
1. 用户调用: provider.CallRpc[MyResp](ctx, "getUser", MyReq{Id: 123})

2. Provider:
   a. 应用 CallOptions (WithTarget? → 选 sender; 否则 → 权重随机)
   b. selectedSender.callRpcRaw(ctx, method, marshaledPayload, headers)

3. RpcSender.callRpcRaw:
   a. requestId = crypto/rand hex(16)
   b. ch = make(chan Packet, 1); pending[requestId] = ch
   c. 构建 Packet{Opcode:Request, Method, Service, Headers{x-sora-rpc-id}, payload}
   d. conn.SendRaw(ctx, pkt) → codec.EncodePacket → transport.Send
   e. select: ch → return | ctx.Done → timeout | senderCtx.Done → stopped

4. 服务端响应 → Connection.readLoop → codec.DecodePacket → handlePacket:
   a. Opcode == Response → OnResponse(conn, pkt)
   b. RpcSender.handleResponse: requestId = pkt.Headers["x-sora-rpc-id"]
   c. 查找 pending[requestId] → ch ← pkt

5. callRpcRaw 收到 pkt:
   a. 删除 pending[requestId]
   b. 返回 pkt 给 Provider

6. Provider.CallRpc:
   a. pkt.Decode[MyResp]() → MyResp 实例
   b. 返回给用户
```

### Endpoint Watch 流程

```
1. Provider.Start():
   a. disco.WatchEndpoints(ctx) → endpointCh
   b. go watchLoop()

2. watchLoop:
   for snapshot := range endpointCh:
     filtered = snapshot.filter(TargetName == serviceName)
     current := senders.keys()
     
     added   = filtered - current   → 为每个新增 endpoint 创建 RpcSender
     removed = current - filtered   → 为每个移除 endpoint 销毁 RpcSender

3. 创建 RpcSender:
   a. 选择 codec: endpoint.Codecs 中第一个已注册的
   b. newRpcSender(endpoint, provider, codec)
   c. sender.Start(ctx) → go connectLoop()
   d. senders[endpointId] = sender
   e. sendersBySvc[endpoint.TargetID] = append(..., sender)

4. 销毁 RpcSender:
   a. sender.Destroy() → cancel ctx, fail pending, disconnect
   b. delete(senders, endpointId)
   c. remove from sendersBySvc
```

### Connection 重连流程

```
connectLoop:
  delay = 500ms
  loop:
    transport = tcp.NewTCPTransport()
    conn = NewConnection(transport, options)
    conn.OnResponse = handleResponse
    
    err = conn.Start(ctx, target, codec)
    if err == nil:
      delay = 500ms  // 重置
      等待 conn 状态变为 Error/Stopped
      failAllPending()
    
    delay = min(delay * 2, 10s)
    select:
      case <-ctx.Done(): return  // endpoint 消失
      case <-time.After(delay): continue
```

## Error Handling

| 场景 | 错误 | 行为 |
|------|------|------|
| 无可用 endpoint | `ErrNoAvailableEndpoint` | CallRpc 直接返回 |
| 指定 serviceId 不存在 | `ErrServiceNotFound` | CallRpc 直接返回 |
| 调用超时 | context 错误 | 删除 pending，返回 |
| Connection 断开 | `ErrConnectionLost` | fail 所有 pending，重连 |
| Endpoint 从 discovery 消失 | `ErrSenderStopped` | Destroy sender，fail pending |
| Provider 已停止 | `ErrProviderStopped` | CallRpc 直接返回 |
| Codec 未找到 | `ErrNoAvailableCodec` | 创建 RpcSender 时报错 |

## Risks / Trade-offs

- **[Breaking: Packet/Codec/Connection API]** 所有使用旧 packet 类型、Codec 接口、Connection 回调的代码需适配——受影响文件明确（connector.go, listener.go, json.go, sora-test/main.go）
- **[Breaking: ListenerCallbacks]** 服务端回调签名变更——OnRequest/OnNotify 参数从具体类型变为 Packet，消费代码需用 `pkt.Decode[T]()` 提取 payload
- **[pending map 泄漏风险]** 如果 response 丢失，pending 条目永不清理——通过 context 超时机制保证最终清理
- **[WatchEndpoints 全量快照]** 每次 discovery 变更推送所有 endpoint，大量 endpoint 时内存/CPU 开销增加——当前规模可接受，未来可优化为增量 watch
- **[connectLoop 竞态]** Destroy 和 connectLoop 并发可能在新连接建立时被取消——通过 ctx cancel 保证安全退出
