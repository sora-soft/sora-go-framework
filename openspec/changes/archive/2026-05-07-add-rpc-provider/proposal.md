## Why

框架当前只有 RPC 接收侧（Listener 接受连接、分发请求），缺少 RPC 发送侧——客户端无法通过发现服务自动连接远端 endpoint，也无法以类型安全的方式发起 RPC 调用。现有 Packet 类型使用 `json.RawMessage` 作为 Payload，与 JSON 格式强耦合，无法适配未来的 Protobuf 等编解码格式。需要构建三层 RPC 发送端框架（Provider → RpcSender → Connection），同时重新设计 Packet 和 Codec 体系以实现编解码无关。

## What Changes

- **BREAKING**: Packet 体系重新设计——删除 `ReqPacketData`、`ResPacketData`、`NotifyPacketData`、`CommandPacketData` 四个独立类型，统一为单一 `Packet` 结构体，Payload 使用私有 `[]byte` + 泛型 `Decode[T]()` 方法隐藏编解码细节
- **BREAKING**: `Codec` 接口重新设计——新增 `Marshal(any)`、`Unmarshal([]byte, any)`、`EncodePacket(Packet)`、`DecodePacket([]byte)` 方法，删除旧的 `Encode(Packet)`/`Decode([]byte)` 方法
- **BREAKING**: `Connection` 回调签名变更——`OnRequest`/`OnNotify` 参数从具体 packet 类型变更为统一 `Packet`，新增 `OnResponse` 回调
- **BREAKING**: `ListenerCallbacks` 签名变更——回调参数从 `*packet.ReqPacketData`/`*packet.NotifyPacketData` 变更为 `Packet`
- 新增 `PayloadCodec` 接口（在 `packet` 包内），定义 `Marshal`/`Unmarshal` 两个方法，供 `Packet.Decode[T]()` 使用
- 新增 `RpcSender`——管理单个 endpoint 的 Connection 生命周期，处理连接/重连/断开，管理 pending request map
- 新增 `Provider`——独立工具，通过 WatchEndpoints 监听服务发现，管理多个 RpcSender，提供权重随机选择和 `CallRpc[Resp any]()` 泛型调用方法
- 新增 `CallOptions`——函数式选项模式，支持 `WithTimeout`、`WithTarget(serviceId)`
- 更新 JSON Codec 适配新 Codec 接口，内部使用 wire struct 维持现有线路格式兼容

## Capabilities

### New Capabilities
- `unified-packet`: 统一的 Packet 类型，支持泛型 `Decode[T]()` 方法，编解码无关的 Payload 管理
- `payload-codec`: `packet.PayloadCodec` 接口定义 Marshal/Unmarshal，供 Packet 内部使用
- `rpc-sender`: RpcSender 管理单个 endpoint 的 Connection，指数退避重连（上限 10s），pending request 管理，连接断开时 fail 所有在途请求
- `rpc-provider`: Provider 通过 discovery.WatchEndpoints 监听远端服务，按 Weight 权重随机选择 endpoint，提供 `CallRpc[Resp any](ctx, method, req)` 泛型调用
- `call-options`: CallOptions 函数式选项，支持 WithTimeout（默认 10s）、WithTarget（指定 ServiceMeta.ID 调用）
- `rpc-id-header`: 使用 `headers["x-sora-rpc-id"]` 关联请求与响应

### Modified Capabilities
- `packet-model`: Packet 从四个独立类型合并为统一类型，新增 PayloadCodec 接口和泛型 Decode 方法
- `connector-lifecycle`: Connection 新增 OnResponse 回调，handlePacket 新增 Response opcode 分发，Send 系列方法适配新 Packet
- `session-management`: ListenerCallbacks 使用统一 Packet 类型
- `codec-interface`: Codec 接口新增 Marshal/Unmarshal/EncodePacket/DecodePacket，删除旧 Encode/Decode

## Impact

- **包结构**: 删除 `packet/req_packet.go`、`packet/response_packet.go`、`packet/notify_packet.go`、`packet/command.go`（合并到 `packet/packet.go`）；新增 `pkg/rpc/provider.go`、`pkg/rpc/rpc_sender.go`、`pkg/rpc/call_options.go`、`pkg/rpc/constants.go`
- **API 签名**: `Codec` 接口、`Connection` 回调、`ListenerCallbacks` 签名变更，所有消费代码需适配
- **线路格式**: JSON 线路格式保持不变（向后兼容），只是内部编解码方式改变
- **依赖**: `packet` 包新增 `PayloadCodec` 接口定义；`pkg/discovery` 的 `WatchEndpoints` 返回类型不变
