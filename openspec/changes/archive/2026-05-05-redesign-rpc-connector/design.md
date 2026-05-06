## Context

sora-go-framework 是一个纯 Go 手写的 RPC 框架，无外部 RPC 依赖（无 gRPC、无 protobuf）。当前 Connector 架构中：

- `Transport` 接口只有 `TCPTransport` 一个实现，职责耦合了连接建立、codec 握手、zlib 压缩和封包
- `Packet` 使用冗余的 getter 接口模式（`ReqPacket`、`ResPacket`、`CommandPacket`），且 Request/Notify 共用 `ReqPacketData`
- 缺乏 context 驱动的超时与重试机制
- 无服务端（Listener）实现，本次不涉及

目录现状：`pkg/rpc/tcp/`、`pkg/rpc/codec/json_buffer_codec.go`

## Goals / Non-Goals

**Goals:**
- Transport 层完全拥有 wire protocol（handshake + 封包/解包），Connector 不参与底层细节
- Connector 通过 `Transport.Connect` 返回值获取服务方确认的 codec，不匹配则进入 Error 状态
- TCPTransport 内部实现可配置的指数退避重试，尊重外部 context 取消
- Packet 模型零 interface，使用 union struct，Request/Notify 拆分为独立类型
- Connector 暴露 `SendRaw(Packet)`，`SendRequest`/`SendCommand` 基于其实现
- 所有 I/O 操作支持 `context.Context` 传播

**Non-Goals:**
- 不实现服务端 Listener/Accept
- 不实现请求-响应关联（Request ID 通过 Header 实现，本次不涉及路由机制）
- 不实现 WebSocket/HTTP 等其他 Transport 实现（只重构接口，预留扩展点）
- 不改变 Ping/Pong 子系统行为

## Decisions

### D1: Transport interface 签名

```go
type Transport interface {
    Connect(ctx context.Context, endpoint string, codec string) (string, error)
    Send(ctx context.Context, data []byte) error
    Recv(ctx context.Context) ([]byte, error)
    Close() error
}
```

**选择**: `Connect` 返回 `(confirmedCodec, error)`。

**替代方案**: `Connect` 返回 `error`，内部验证 codec 匹配。被否决——Connector 需要知道确认结果来切换状态。

**理由**: Connector 从 Connecting(pending) 到 Ready 的状态转换依赖于收到服务方确认。返回值让 Connector 显式感知 handshake 结果。

### D2: Packet union struct

```go
type Packet struct {
    Opcode  PacketOpcode
    Req     *ReqPacketData
    Notify  *NotifyPacketData
    Res     *ResPacketData
    Cmd     *CommandPacketData
}
```

**选择**: 零 interface，使用 tagged union struct。

**替代方案**: 保留 `Packet interface { GetOpcode() }` + 具体 struct。被否决——用户要求去掉所有 getter 接口。

**理由**: 消除 getter 接口冗余（导出字段 + getter 方法同时存在），Codec.Encode/Decode 通过 Opcode 分支处理，调用方通过字段直接访问。每个 Opcode 对应且仅对应一个非 nil 指针字段。

### D3: Request / Notify 拆分

`ReqPacketData`(opcode=1) 和 `NotifyPacketData`(opcode=3) 字段相同但为独立 struct。

**理由**: 语义不同（Request 期待 Response，Notify 是单向），独立类型允许未来各自演进而不互相影响。

### D4: TCPTransport 内部重试

```go
type TCPOptions struct {
    MaxRetries     int           // default: 3
    InitialDelay   time.Duration // default: 500ms
    MaxDelay       time.Duration // default: 8s
    ConnectTimeout time.Duration // per-attempt, default: 5s
}
```

重试策略：指数退避 `delay = min(InitialDelay * 2^attempt, MaxDelay)`。每次重试前检查 `ctx.Err()`，等待期间通过 `select` 监听 `ctx.Done()`。

**理由**: 重试是 TCP 层的关注点（网络抖动、连接拒绝），Connector 不应感知。指数退避防止服务端过载时雪崩。

### D5: Codec interface

```go
type Codec interface {
    GetCode() string
    Encode(pkt Packet) ([]byte, error)
    Decode(data []byte) (Packet, error)
}
```

**选择**: Encode/Decode 使用 `Packet` union struct 而非 interface。

### D6: Connector 暴露方法

```go
func (c *Connector) SendRaw(ctx context.Context, pkt Packet) error
```

`SendRequest` 和 `SendCommand` 作为便利方法调用 `SendRaw`。

### D7: 目录结构

```
pkg/rpc/
├── connector.go
├── options.go
├── transport.go
├── codec.go
├── pinger.go
├── listener.go
├── packet/
│   ├── packet.go           Packet union struct + Opcode
│   ├── req_packet.go       ReqPacketData
│   ├── notify_packet.go    NotifyPacketData
│   ├── response_packet.go  ResPacketData
│   └── command.go          CommandPacketData + ConnectorCommand
├── codec/
│   └── json/
│       └── json.go         JSONBufferCodec
└── transport/
    └── tcp/
        └── tcp.go          TCPTransport + retry + handshake + zlib
```

## Risks / Trade-offs

- **[BREAKING API]** 所有接口签名变更 → 调用方（`cmd/sora-test`）需完全重写连接代码
- **[Union struct 内存]** `Packet` struct 包含 4 个指针字段，任意时刻仅 1 个非 nil → 微量内存浪费，但对 RPC 包大小可忽略
- **[Codec 实现复杂度]** JSONCodec.Decode 需要两阶段反序列化（先读 opcode，再按类型解具体 struct）→ 与当前实现一致，无新增复杂度
- **[无 Listener 端]** Transport.Connect 只定义了发起方行为，服务端需要对应的 Accept 逻辑 → 暂不实现，但接口设计需预留服务端对称性
