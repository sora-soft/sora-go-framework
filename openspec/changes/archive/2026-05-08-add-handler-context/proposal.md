## 为什么

当前 RPC Router 的类型安全 handler（`Method` 和 `Notify`）只接收解码后的请求体，无法读取请求 Headers（如认证 token、trace-id）或设置响应 Headers。Middleware 可以访问原始 Packet，但 handler 与 middleware 之间没有数据传递通道。这导致认证、链路追踪等横切关注点无法在 handler 层自然地实现。

## 变更内容

- **新增 `HandlerContext` 接口**：定义 handler 可用的通用请求上下文 API（读取 headers、获取 method/service、middleware→handler 传值）
- **新增 `RequestContext` 结构体**：用于 Method handler，包含 `RequestReader`（只读请求侧）和 `ResponseWriter`（只写响应侧），支持读写 headers 和 middleware 传值
- **新增 `NotifyContext` 结构体**：用于 Notify handler，仅包含 `RequestReader`，无响应写入能力
- **新增 `RequestReader`**：封装请求元数据的只读访问（Header、Headers、Method、Service）
- **新增 `ResponseWriter`**：封装响应头的写入（SetHeader、SetHeaders）
- **BREAKING**：`Method` handler 签名从 `func(conn *Connection, req Req) (Resp, error)` 变更为 `func(ctx *RequestContext, req Req) (Resp, error)`
- **BREAKING**：`Notify` handler 签名从 `func(conn *Connection, msg Msg) error` 变更为 `func(ctx *NotifyContext, msg Msg) error`
- **BREAKING**：`DispatchFunc` 签名从 `func(conn *Connection, pkt packet.Packet) error` 变更为 `func(ctx HandlerContext) error`
- **BREAKING**：`Middleware` 签名随 `DispatchFunc` 一同变更
- **响应头合并**：handler 通过 `ResponseWriter` 设置的响应头与请求头合并后发出（handler 设置的同名 key 覆盖请求头）

## 功能 (Capabilities)

### 新增功能
- `handler-context`: Handler 请求上下文系统——HandlerContext 接口、RequestContext/NotifyContext 结构体、RequestReader/ResponseWriter，为 handler 提供请求头读取、响应头设置和 middleware→handler 数据传递能力

### 修改功能
- `rpc-router`: Router 的 Method/Notify handler 签名、DispatchFunc/Middleware 签名、以及内部 dispatch 流程需要适配新的 Context 体系

## 影响

- **API 破坏性变更**：所有使用 `router.Method`、`router.Notify`、`router.Use` 的代码需要更新 handler 签名
- **核心文件**：`pkg/rpc/router/router.go`（dispatch 逻辑重构）、新增 `pkg/rpc/context.go`、`pkg/rpc/reader.go`、`pkg/rpc/writer.go`
- **示例代码**：`cmd/sora-test/main.go` 需要更新
- **客户端不受影响**：`pkg/rpc/provider/*` 无需变更，此变更仅影响服务端 handler 侧
