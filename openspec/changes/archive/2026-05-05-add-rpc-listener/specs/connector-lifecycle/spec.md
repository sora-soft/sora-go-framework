## MODIFIED Requirements

### Requirement: Connector 状态机
Connector SHALL 实现状态机：`Init(1) → Connecting(2) → Ready(4) → Stopping(5) → Stopped(6)`，任意状态可进入 `Error(100)`。状态迁移 SHALL 使用已有的 `utility.LifeCycle[T]`。Client-side 通过 `Start` 进入 Connecting，Server-side 通过 `Serve` 进入 Connecting。

#### Scenario: 正常连接流程（client）
- **WHEN** 调用 `Start(ctx, target, codec)` 且 Transport.Connect 成功
- **THEN** 状态依次为 Init → Connecting → Ready

#### Scenario: 连接失败（client）
- **WHEN** 调用 `Start(ctx, target, codec)` 且 Transport.Connect 返回 error
- **THEN** 状态为 Error

#### Scenario: codec 不匹配（client）
- **WHEN** Transport.Connect 返回的 confirmedCodec 与 codec.GetCode() 不一致
- **THEN** 状态为 Error

#### Scenario: 正常断开
- **WHEN** 调用 `Disconnect()` 且当前状态为 Ready
- **THEN** 状态依次为 Ready → Stopping → Stopped

#### Scenario: 服务端初始化流程（server）
- **WHEN** 调用 `Serve()` 且 Transport.Handshake 成功且 codec 有效
- **THEN** 状态依次为 Init → Connecting → Ready

#### Scenario: 服务端 handshake 失败（server）
- **WHEN** 调用 `Serve()` 且 Transport.Handshake 返回 error
- **THEN** 状态为 Error

#### Scenario: 服务端 codec 不支持（server）
- **WHEN** 调用 `Serve()` 且 Handshake 返回的 codec 在全局 registry 中不存在
- **THEN** 状态为 Error

## ADDED Requirements

### Requirement: Connection.Serve server-side 初始化
Connection SHALL 提供 `Serve() error` 方法作为 server-side 初始化入口。Serve 调用 `Transport.Handshake` 获取客户端选择的 codec，通过全局 registry 查找验证，成功后启动 readLoop 并进入 Ready。

#### Scenario: Serve 完整流程
- **WHEN** 调用 `Serve()` 且 Transport.Handshake 返回 `"json"` 且 GetCodec("json") 成功
- **THEN** 设置 codec，状态为 Init → Connecting → Ready，启动 readLoop

#### Scenario: Serve codec 不支持
- **WHEN** 调用 `Serve()` 且 Handshake 返回的 codec 在全局 registry 中不存在
- **THEN** 状态为 Error，返回 codec 不支持的错误

### Requirement: Connection.SendResponse
Connection SHALL 提供 `SendResponse(ctx context.Context, res *packet.ResPacketData) error` 便利方法，构造 `Packet{Opcode: PacketOpcodeResponse, Res: res}` 后调用 `SendRaw`。

#### Scenario: SendResponse 实现
- **WHEN** 调用 `SendResponse(ctx, &ResPacketData{...})`
- **THEN** 构造 `Packet{Opcode: PacketOpcodeResponse, Res: res}` 并调用 `SendRaw`

### Requirement: Connector 暴露 SendRaw 方法
Connector SHALL 暴露 `SendRaw(ctx context.Context, pkt Packet) error` 方法，将 Packet 通过 Codec.Encode 序列化后通过 Transport.Send 发送。

#### Scenario: SendRaw 发送 Request
- **WHEN** 调用 `SendRaw(ctx, Packet{Opcode: 1, Req: &ReqPacketData{...}})`
- **THEN** Codec.Encode 序列化 Packet，Transport.Send 发送序列化后的 []byte

#### Scenario: SendRaw 发送 Command
- **WHEN** 调用 `SendRaw(ctx, Packet{Opcode: 4, Cmd: &CommandPacketData{...}})`
- **THEN** Codec.Encode 序列化 Packet，Transport.Send 发送序列化后的 []byte

### Requirement: SendRequest 和 SendCommand 基于 SendRaw
`SendRequest` 和 `SendCommand` SHALL 作为便利方法实现，内部构造 `Packet` union struct 后调用 `SendRaw`。

#### Scenario: SendRequest 实现
- **WHEN** 调用 `SendRequest(ctx, &ReqPacketData{...})`
- **THEN** 构造 `Packet{Opcode: PacketOpcodeRequest, Req: req}` 并调用 `SendRaw`

#### Scenario: SendCommand 实现
- **WHEN** 调用 `SendCommand(ctx, &CommandPacketData{...})`
- **THEN** 构造 `Packet{Opcode: PacketOpcodeCommand, Cmd: cmd}` 并调用 `SendRaw`

### Requirement: Connector 所有操作支持 context
Connector 的 `Start`、`Serve`、`SendRaw`、`Disconnect` SHALL 接受 `context.Context` 并传播至 Transport 层。Connector SHALL 持有一个生命周期 context，`Disconnect` 时取消该 context。

#### Scenario: Start 期间 context 取消
- **WHEN** 调用 `Start(ctx, ...)` 且 ctx 在 Transport.Connect 过程中被取消
- **THEN** Start 返回 context 错误，状态为 Error

#### Scenario: Disconnect 取消生命周期 context
- **WHEN** 调用 `Disconnect()`
- **THEN** 取消内部生命周期 context，readLoop 终止

### Requirement: readLoop 使用 Packet union struct
Connector 的 readLoop SHALL 通过 `Transport.Recv` 接收 []byte，通过 `Codec.Decode` 反序列化为 `Packet` struct，根据 `Opcode` 分发处理。

#### Scenario: readLoop 收到 Command packet
- **WHEN** readLoop 收到数据并 Decode 得到 `Packet{Opcode: 4, Cmd: &CommandPacketData{Command: "ping"}}`
- **THEN** 调用 handleCommand 处理 ping

#### Scenario: readLoop 收到错误
- **WHEN** readLoop 中 `Recv` 或 `Decode` 返回错误
- **THEN** 状态切换为 Error
