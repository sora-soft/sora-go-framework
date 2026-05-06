## Why

当前 RPC Connector 架构存在三个核心问题：(1) Transport 层职责不清晰——TCPTransport 同时耦合了连接建立、codec 握手协商、zlib 压缩和 length-prefix 封包；(2) Packet 模型使用冗余的 getter 接口且 Request/Notify 共享同一个 struct，语义模糊；(3) 缺乏 context 驱动的超时与重试机制。为了支持多种底层传输协议（TCP、WebSocket、HTTP 等）的灵活扩展，需要重构整个 Connector 架构。

## What Changes

- **BREAKING**: 重构 `Transport` interface，`Connect` 返回 `(string, error)` 表示服务方确认的 codec 名称
- **BREAKING**: 重构 `Packet` 模型，去掉所有 getter 接口（`ReqPacket`、`ResPacket`、`CommandPacket`），改用 union struct `Packet`
- **BREAKING**: 拆分 `ReqPacketData` 为独立的 `ReqPacketData`（opcode=1）和 `NotifyPacketData`（opcode=3）
- **BREAKING**: 重构 `Codec` interface，Encode/Decode 使用新的 `Packet` union struct
- 重构 `TCPTransport`：内含指数退避重试（可配置）、handshake 握手协商、zlib+length-prefix 封包
- 重构 `Connector`：暴露 `SendRaw(Packet)` 方法，`SendRequest`/`SendCommand` 基于其实现
- 所有操作支持 `context.Context` 传播，重试逻辑尊重外部 context 取消
- 目录结构调整：`tcp/` → `transport/tcp/`，`codec/json_buffer_codec.go` → `codec/json/json.go`

## Capabilities

### New Capabilities
- `transport-interface`: 可插拔的传输层接口，支持灵活选择底层通讯协议（TCP/WebSocket/等），Transport 完全拥有 wire protocol（handshake + 封包/解包）
- `packet-model`: 基于 union struct 的 Packet 模型，零 interface，独立 Request/Notify/Response/Command 类型
- `connector-lifecycle`: Connector 生命周期管理，context 驱动的超时控制，SendRaw 暴露

### Modified Capabilities

## Impact

- **BREAKING API**: `Transport`、`Codec`、`Packet` 接口全部变更，所有依赖方需适配
- **BREAKING API**: `NewConnection` 和 `Connection.Start` 签名变更
- 目录结构变更：`pkg/rpc/tcp/` → `pkg/rpc/transport/tcp/`，`pkg/rpc/codec/` 下文件重组
- `cmd/sora-test/main.go` 需适配新 API
- 无外部依赖变更（仍然只依赖 uuid + x/sync）
