## 1. Packet 模型重构

- [x] 1.1 创建 `pkg/rpc/packet/packet.go`：定义 `PacketOpcode` 枚举常量（1=Request, 2=Response, 3=Notify, 4=Command）和 `Packet` union struct（Opcode + Req/Notify/Res/Cmd 四个指针字段）
- [x] 1.2 重构 `pkg/rpc/packet/req_packet.go`：`ReqPacketData` 仅用于 opcode=1，删除 getter 方法
- [x] 1.3 创建 `pkg/rpc/packet/notify_packet.go`：`NotifyPacketData` 用于 opcode=3，字段与 ReqPacketData 相同（Method, Service, Headers, Payload）
- [x] 1.4 重构 `pkg/rpc/packet/response_packet.go`：`ResPacketData` 删除 getter 方法，保留 Opcode(=2)、Headers、Payload 字段
- [x] 1.5 重构 `pkg/rpc/packet/command.go`：`CommandPacketData` 删除 getter 方法，保留 Opcode(=4)、Command、Args 字段；保留 `ConnectorCommand` 常量和构造函数
- [x] 1.6 删除 `pkg/rpc/packet/interface.go`：移除所有 getter interface（Packet/ReqPacket/ResPacket/CommandPacket）

## 2. Transport 接口重构

- [x] 2.1 重构 `pkg/rpc/transport.go`：定义新 `Transport` interface，`Connect(ctx, endpoint, codec) → (string, error)`，`Send(ctx, []byte) → error`，`Recv(ctx) → ([]byte, error)`，`Close() → error`

## 3. Codec 接口重构

- [x] 3.1 重构 `pkg/rpc/interface.go`：`Codec` interface 改为 `GetCode() string`，`Encode(Packet) → ([]byte, error)`，`Decode([]byte) → (Packet, error)`
- [x] 3.2 重构 `pkg/rpc/codec.go`：更新全局 codec registry 类型签名以适配新 Codec interface
- [x] 3.3 创建 `pkg/rpc/codec/json/json.go`：JSONBufferCodec 适配新 Packet union struct，Encode 按 Opcode 分支序列化，Decode 两阶段反序列化（先读 opcode 再按类型解）
- [x] 3.4 删除旧的 `pkg/rpc/codec/json_buffer_codec.go`

## 4. TCPTransport 重构

- [x] 4.1 创建 `pkg/rpc/transport/tcp/` 目录
- [x] 4.2 实现 `TCPOptions` struct：MaxRetries(default 3)、InitialDelay(default 500ms)、MaxDelay(default 8s)、ConnectTimeout(default 5s)
- [x] 4.3 实现 `TCPTransport.Connect`：TCP dial + 发送 `"${codec}\n"` + 接收确认 + 验证匹配 + 返回 `(confirmedCodec, error)`
- [x] 4.4 实现指数退避重试：循环 MaxRetries 次，delay = min(InitialDelay * 2^attempt, MaxDelay)，等待期间 select 监听 ctx.Done()
- [x] 4.5 实现 `TCPTransport.Send`：zlib 压缩 + 4字节 big-endian 长度头 + 发送
- [x] 4.6 实现 `TCPTransport.Recv`：读取 4字节长度头 + 读取压缩数据 + zlib 解压
- [x] 4.7 实现 `TCPTransport.Close`：关闭底层 net.Conn
- [x] 4.8 删除旧的 `pkg/rpc/tcp/` 目录

## 5. Connector 重构

- [x] 5.1 创建 `pkg/rpc/options.go`：`ConnectorOptions` 包含 Ping 配置
- [x] 5.2 重构 `pkg/rpc/connector.go`：`Connector` struct 使用新 Transport 和 Codec 类型，保留 LifeCycle 状态机
- [x] 5.3 实现 `Start(ctx, target, codec)`：调用 `Transport.Connect`，验证 confirmedCodec 匹配，状态 Init → Connecting → Ready，启动 readLoop + pinger
- [x] 5.4 实现 `readLoop`：Recv → Decode(Packet union struct) → 按 Opcode 分发处理
- [x] 5.5 实现 `SendRaw(ctx, Packet) error`：Codec.Encode → Transport.Send
- [x] 5.6 实现 `SendRequest(ctx, *ReqPacketData)` 和 `SendCommand(ctx, *CommandPacketData)` 基于 SendRaw
- [x] 5.7 适配 handleCommand/handlePacket 使用新 Packet union struct
- [x] 5.8 实现 `Disconnect()`：取消生命周期 context，状态 Ready → Stopping → Stopped

## 6. Pinger 适配

- [x] 6.1 重构 `pkg/rpc/pinger.go`：适配新 Packet union struct 和 CommandPacketData（直接访问字段而非 getter）

## 7. 集成验证

- [x] 7.1 更新 `cmd/sora-test/main.go` 适配新 API（新 Transport、Codec、Packet 构造方式）
- [x] 7.2 验证编译通过
