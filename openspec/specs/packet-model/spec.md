### Requirement: Packet union struct 替代 interface
系统 SHALL 使用 `Packet` struct 作为所有 packet 类型的统一容器，替代原有 `Packet` interface。`Packet` struct SHALL 包含 `Opcode` 字段和四个指针字段：`Req`、`Notify`、`Res`、`Cmd`。任意时刻仅有一个指针字段非 nil，由 Opcode 决定。

#### Scenario: 构造 Request packet
- **WHEN** 创建一个 opcode=1 的 packet
- **THEN** `Packet.Opcode == PacketOpcodeRequest`，`Packet.Req` 非 nil，其余指针字段为 nil

#### Scenario: 构造 Notify packet
- **WHEN** 创建一个 opcode=3 的 packet
- **THEN** `Packet.Opcode == PacketOpcodeNotify`，`Packet.Notify` 非 nil，其余指针字段为 nil

#### Scenario: 构造 Response packet
- **WHEN** 创建一个 opcode=2 的 packet
- **THEN** `Packet.Opcode == PacketOpcodeResponse`，`Packet.Res` 非 nil，其余指针字段为 nil

#### Scenario: 构造 Command packet
- **WHEN** 创建一个 opcode=4 的 packet
- **THEN** `Packet.Opcode == PacketOpcodeCommand`，`Packet.Cmd` 非 nil，其余指针字段为 nil

### Requirement: 独立的 Request 和 Notify struct
系统 SHALL 提供 `ReqPacketData`(opcode=1) 和 `NotifyPacketData`(opcode=3) 两个独立 struct。两者字段相同（Method, Service, Headers, Payload）但类型独立。

#### Scenario: ReqPacketData 序列化
- **WHEN** 序列化 `ReqPacketData`（Method="GetUser", Service="UserService"）
- **THEN** JSON 输出包含 `"opcode": 1`，`"method": "GetUser"`，`"service": "UserService"`

#### Scenario: NotifyPacketData 序列化
- **WHEN** 序列化 `NotifyPacketData`（Method="OnEvent", Service="EventService"）
- **THEN** JSON 输出包含 `"opcode": 3`，`"method": "OnEvent"`，`"service": "EventService"`

### Requirement: 去掉所有 getter interface
系统 SHALL NOT 包含 `ReqPacket`、`ResPacket`、`CommandPacket` getter interfaces。所有 packet 类型访问通过 struct 字段直接访问。

#### Scenario: 访问 Request 的 Method
- **WHEN** 持有 `Packet` 且 `Opcode == PacketOpcodeRequest`
- **THEN** 通过 `pkt.Req.Method` 直接访问 Method 字段

### Requirement: Codec 使用 Packet union struct
`Codec.Encode` SHALL 接受 `Packet` struct，根据 `Opcode` 分支序列化对应的指针字段。`Codec.Decode` SHALL 返回 `Packet` struct，先解码 opcode，再按类型反序列化。

#### Scenario: JSONCodec Encode Request
- **WHEN** 调用 `Encode(Packet{Opcode: 1, Req: &ReqPacketData{...}})`
- **THEN** 序列化 `pkt.Req` 为 JSON

#### Scenario: JSONCodec Decode Request
- **WHEN** 调用 `Decode([]byte)` 且 JSON 包含 `"opcode": 1`
- **THEN** 返回 `Packet{Opcode: 1, Req: &ReqPacketData{...}}`，Req 字段已填充

### Requirement: CommandPacketData 保留在 packet 层
`CommandPacketData` 和 `ConnectorCommand` 常量 SHALL 继续定义在 `packet` 包中。

#### Scenario: Command packet 结构
- **WHEN** 创建 CommandPacketData
- **THEN** 包含 Opcode(=4)、Command(string)、Args(json.RawMessage) 三个导出字段
