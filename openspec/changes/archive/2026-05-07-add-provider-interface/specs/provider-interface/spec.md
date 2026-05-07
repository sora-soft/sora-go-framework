## ADDED Requirements

### 需求:Provider 接口定义
系统必须在 `pkg/rpc/provider` 包中定义 `Provider` interface，包含以下方法：
- `Start(ctx context.Context) error`
- `Stop() error`
- `CallRpc(ctx context.Context, method string, req any, opts ...CallOption) (packet.Packet, error)`

#### 场景:interface 满足
- **当** 定义了 Provider interface
- **那么** rpcProvider（具体实现）必须自动满足该 interface

### 需求:Provider 具体实现隐藏
系统必须将当前导出的 `Provider` struct 重命名为不导出的 `rpcProvider`。`NewProvider()` 函数必须返回 `Provider` interface 类型。

#### 场景:外部通过 interface 使用
- **当** 外部代码调用 `provider.NewProvider()`
- **那么** 返回值为 `Provider` interface，外部代码可以调用 Start、Stop、CallRpc 方法

#### 场景:无法访问具体类型
- **当** 外部代码尝试引用 `provider.rpcProvider`
- **那么** 编译必须失败（不导出类型）

### 需求:RpcSender 清理
系统必须从 `RpcSender` struct 中删除未使用的 `provider` 字段。`NewRpcSender` 函数签名必须移除 `provider` 参数。

#### 场景:RpcSender 不再持有 provider 引用
- **当** 创建新的 RpcSender 实例
- **那么** 不再需要传入 Provider 参数，RpcSender 不持有 Provider 引用

### 需求:引用计数保持不变
rpcProvider 必须保持现有的引用计数（refCount）机制：Start 增加计数，Stop 减少计数，仅当计数降为 0 时执行实际启停。

#### 场景:多次注册同一 Provider
- **当** 同一个 rpcProvider 实例被 Start 多次
- **那么** refCount 递增，不重复启动 watchLoop

#### 场景:最后一次 Stop
- **当** refCount 降为 0
- **那么** 执行实际关闭（cancel context、清理 senders）
