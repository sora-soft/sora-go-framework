## Why

当前 `ListenerCallbacks` 的 `OnRequest` / `OnNotify` 是扁平回调，所有 method 的请求汇入同一个函数，使用者需要自行 decode payload、按 `pkt.Method` 手动分发、构造 response。随着 method 增多，回调内会堆积大量 `switch/case` 和重复的 encode/decode 样板代码，且缺乏统一的错误处理和中间件支持。

需要一个 Router 层封装 `ListenerCallbacks`，提供基于 `Method` 的一级路由、强类型 handler 注册、自动 request/response 编解码、以及 middleware 链。

## What Changes

- **New**: `Router` —— 包裹 `ListenerCallbacks` 的路由层，提供 `Method[Req, Resp]` 和 `Notify[Msg]` 两个泛型注册方法
- **New**: 中间件支持 `router.Use(middleware)`，在路由分发前执行（日志、鉴权等）
- **New**: 自动 Response 封装，遵循 `{error: IPayloadError | null, result: T | null}` 格式
- **New**: 统一错误处理 —— handler 返回 `*errorx.Error` 自动映射，普通 error 兜底，未注册 method 返回 `ERR_METHOD_NOT_FOUND`
- Router 同时服务于 Listener 侧和 Connector 侧的 `OnRequest` / `OnNotify` 回调

## Capabilities

### New Capabilities

- `rpc-router`: 基于 Method 的一级路由分发，泛型强类型 handler 注册（`Method[Req, Resp]` / `Notify[Msg]`），middleware 链，自动 response 编解码与错误封装

### Modified Capabilities

（无，Router 为纯包装层，不修改现有 `Listener` / `Connection` 代码）

## Impact

- **新增代码**: `pkg/rpc/router/` 包（Router、middleware 类型、response 封装逻辑）
- **现有代码**: 不修改 `Listener`、`Connection`、`ListenerCallbacks` 任何代码
- **使用侧**: 新代码可选使用 Router 包装 `ListenerCallbacks`；旧代码无需改动
- **依赖**: 仅依赖现有 `pkg/rpc/packet` 和 `pkg/utility/errorx`，无外部依赖引入
