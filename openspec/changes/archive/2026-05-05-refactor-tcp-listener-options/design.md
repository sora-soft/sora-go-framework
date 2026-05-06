## Context

当前 `TCPListener` 通过 `net.Listen("tcp", addr)` 绑定固定地址，端口由调用方硬编码。多实例部署时共享同一份配置文件，所有实例绑定同一端口导致冲突。

`TCPOptions` 同时服务于 `TCPListener`（服务端）和 `TCPTransport`（客户端），但两者的关注点完全不同——listener 关心绑定地址，transport 关心重连策略。

## Goals / Non-Goals

**Goals:**
- 支持端口范围扫描，多实例启动时自动选择可用端口
- 支持 `Port`（固定）和 `PortRange`（范围）两种模式，二选一互斥
- 简化 API：移除 `TCPOptions`，listener 和 transport 各自管理自己的配置

**Non-Goals:**
- 不支持端口段自动扩容
- 不提供端口扫描的高级策略（如指数退避、并发探测）
- 不改变 `TCPTransport` 的重连/退避逻辑本身，仅硬编码原参数

## Decisions

### 1. 端口扫描策略：单调递增 + 随机步进

从 `PortRange[0]` 开始，每次绑定失败后 `current += rand(1, 5)`，直到成功或超出 `PortRange[1]`。

**为什么不是线性扫描或纯随机：**
- 线性扫描：多实例同时启动时全部从同一端口开始，碰撞概率高
- 纯随机：可能重复访问已尝试的端口
- 随机步进：单调递增保证不重复，随机性分散不同实例的扫描路径

### 2. 移除 TCPOptions，硬编码传输参数

`TCPTransport` 的 `MaxRetries`、`InitialDelay`、`MaxDelay`、`ConnectTimeout` 当前无外部配置需求，硬编码为默认值（3 / 500ms / 8s / 5s）。

### 3. TCPListenerOptions 替代 addr string

```go
type TCPListenerOptions struct {
    Host      string
    Port      int     // 固定端口模式
    PortRange []int   // [min, max] 范围扫描模式
}
```

`Port` 和 `PortRange` 二选一，同时设置或都不设置视为配置错误。

### 4. 签名变更

```go
// 之前
func NewTCPListener(addr string, opts TCPOptions, connOpts rpc.ConnectorOptions) (*TCPListener, error)

// 之后
func NewTCPListener(opts TCPListenerOptions, connOpts rpc.ConnectorOptions) (*TCPListener, error)
```

## Risks / Trade-offs

- **[BREAKING API]** 所有 `NewTCPListener` 调用方需适配 → 搜索所有引用并更新
- **[硬编码传输参数]** 未来如需动态调整 transport 参数需再次重构 → 当前无此需求，可接受
- **[随机步进可能跳过可用端口]** 步进 1-5 可能跳过中间的可用端口 → range 足够大时影响可忽略
