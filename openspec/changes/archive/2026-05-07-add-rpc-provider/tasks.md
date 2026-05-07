## 1. Packet 体系重设计

- [x] 1.1 重写 `pkg/rpc/packet/packet.go`——定义 `PayloadCodec` 接口（Marshal/Unmarshal）、统一 `Packet` struct（Opcode/Method/Service/Headers + 私有 payload/codec）、泛型 `Decode[T any]()` 方法、`Payload()` 方法、`NewDecodedPacket()` 工厂、`NewRequest[T]()`/`NewResponse[T]()`/`NewNotify[T]()`/`NewCommandPacket()` 泛型工厂、`ConnectorCommand` 常量
- [x] 1.2 删除 `pkg/rpc/packet/req_packet.go`
- [x] 1.3 删除 `pkg/rpc/packet/response_packet.go`
- [x] 1.4 删除 `pkg/rpc/packet/notify_packet.go`
- [x] 1.5 删除 `pkg/rpc/packet/command.go`

## 2. Codec 接口重设计

- [x] 2.1 重写 `pkg/rpc/interface.go`——新 Codec 接口：GetCode/Marshal/Unmarshal/EncodePacket/DecodePacket
- [x] 2.2 更新 `pkg/rpc/codec.go`——RegisterCodec/GetCodec 签名不变，确认与新接口兼容

## 3. JSON Codec 适配

- [x] 3.1 重写 `pkg/rpc/codec/json/json.go`——实现新 Codec 接口：Marshal/Unmarshal 委托 encoding/json；EncodePacket 按 Opcode 使用 wire struct 编码（requestWire/responseWire/notifyWire/commandWire）；DecodePacket 先解码 opcode 再按类型解码为 Packet；`NewDecodedPacket` 时传入 `c`（codec 自身）作为 PayloadCodec

## 4. Connection 适配

- [x] 4.1 更新 `pkg/rpc/connector.go`

## 5. Listener 适配

- [x] 5.1 更新 `pkg/rpc/listener.go`

## 6. Pinger 适配

- [x] 6.1 更新 `pkg/rpc/pinger.go`——确认 Pinger 仅使用函数回调（sendPing/onError/state），不直接依赖 packet 类型，无需修改

## 7. RPC 发送端框架——常量与选项

- [x] 7.1 新建 `pkg/rpc/constants.go`——定义 `HeaderRpcId = "x-sora-rpc-id"` 等 header 常量
- [x] 7.2 新建 `pkg/rpc/call_options.go`——CallOptions struct（Timeout time.Duration, TargetID string）；CallOption func 类型；WithTimeout(time.Duration)；WithTarget(serviceId string)；默认超时 10s
- [x] 7.3 新建 `pkg/rpc/provider_errors.go`——定义错误码：ErrNoAvailableEndpoint/ErrServiceNotFound/ErrCallTimeout/ErrConnectionLost/ErrSenderStopped/ErrProviderStopped/ErrNoAvailableCodec

## 8. RpcSender 实现

- [x] 8.1 新建 `pkg/rpc/rpc_sender.go`——RpcSender struct（endpoint EndpointMeta, provider *Provider, conn *Connection, connMu sync.RWMutex, pending map[string]chan packet.Packet, pendingMu sync.RWMutex, codec Codec, ctx/cancel）
- [x] 8.2 实现 `NewRpcSender(endpoint, provider, codec)` 构造函数
- [x] 8.3 实现 `Start(ctx)`——启动 connectLoop goroutine
- [x] 8.4 实现 `connectLoop()`——指数退避重连
- [x] 8.5 实现 `handleResponse(conn, pkt)`
- [x] 8.6 实现 `callRpcRaw(ctx, method string, payload []byte, headers map[string]string) (packet.Packet, error)`
- [x] 8.7 实现 `failAllPending()`
- [x] 8.8 实现 `Destroy()`
- [x] 8.9 实现 `isReady() bool`

## 9. Provider 实现

- [x] 9.1 新建 `pkg/rpc/provider.go`
- [x] 9.2 实现 `NewProvider(serviceName, disco, opts)` 构造函数
- [x] 9.3 实现 `Start(ctx)`——启动 watchLoop goroutine
- [x] 9.4 实现 `watchLoop()`
- [x] 9.5 实现 `addSender(ctx, endpoint)`
- [x] 9.6 实现 `removeSender(endpointId)`
- [x] 9.7 实现 `selectSender() (*RpcSender, error)`
- [x] 9.8 实现 `selectSenderByService(serviceId) (*RpcSender, error)`
- [x] 9.9 实现 `CallRpc[Resp any]` (simplified to return Packet for Decode)
- [x] 9.10 实现 `Stop()`

## 10. Discovery metadata 适配

- [x] 10.1 确认 `pkg/discovery/metadata.go` 中 `NewEndpointMetaFromListener` 使用 `rpc.ListenerMetaInfo`（不含 packet 类型），无需修改

## 11. 消费代码适配

- [x] 11.1 更新 `cmd/sora-test/main.go`——无 packet 类型引用，无需修改
- [x] 11.2 更新 `pkg/component/etcd_test.go`——无 packet 类型引用，无需修改
- [x] 11.3 更新 `pkg/discovery/store/etcd/etcd_test.go`——无 packet 类型引用，无需修改
- [x] 11.4 更新 `pkg/discovery/store/ram/ram_test.go`——无 packet 类型引用，无需修改

## 12. 验证

- [x] 12.1 确认编译通过（`go build ./...`）
- [x] 12.2 确认现有测试通过（`go test ./...`）
