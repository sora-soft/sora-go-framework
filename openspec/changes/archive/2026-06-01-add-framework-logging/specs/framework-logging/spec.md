## ADDED Requirements

### 需求:Runtime 启动流程日志
Runtime 启动时必须依次输出以下日志事件：
- `load-config` (Info): 配置加载完成，content 包含 config 字段
- `connect-discovery` 成功后输出 (Info): discovery 连接信息
- `install-node` 成功后输出 (Info): node 安装信息
- `start-runtime-success` (Success): 启动完成，content 包含 discovery 和 node 信息
- `connect-discovery` 失败时必须输出 (Fatal) 日志
- `install-node` 失败时必须输出 (Fatal) 日志
- `register-node` 失败时必须输出 (Fatal) 日志

#### 场景:Runtime 正常启动
- **当** 调用 `Runtime.Startup()` 且 discovery 连接成功、node 安装成功
- **那么** 依次输出 load-config、connect-discovery、install-node、start-runtime-success 日志，级别为 Info/Info/Info/Success

#### 场景:Discovery 连接失败
- **当** 调用 `Runtime.Startup()` 且 discovery 连接失败
- **那么** 输出 Fatal 级别日志，event 为 `connect-discovery`

#### 场景:Node 安装失败
- **当** 调用 `Runtime.Startup()` 且 node 安装失败
- **那么** 输出 Fatal 级别日志，event 为 `install-node`

### 需求:Runtime 关闭流程日志
Runtime 关闭时必须输出以下日志事件：
- `process-command` (Info): 收到 SIGINT/SIGTERM 时，content 包含 command 字段
- `service-stopping` / `service-stopped` (Info/Success): 每个服务卸载
- `worker-stopping` / `worker-stopped` (Info/Success): 每个 Worker 卸载
- `all-service-closed` (Info): 所有服务关闭完成
- `all-worker-closed` (Info): 所有 Worker 关闭完成
- `discovery-disconnected` (Info): Discovery 断开连接

#### 场景:收到 SIGINT 信号
- **当** 进程收到 SIGINT 信号
- **那么** 输出 Info 级别日志，event 为 `process-command`，command 为 `SIGINT`，随后触发 Shutdown

#### 场景:收到 SIGTERM 信号
- **当** 进程收到 SIGTERM 信号
- **那么** 输出 Info 级别日志，event 为 `process-command`，command 为 `SIGTERM`，随后触发 Shutdown

#### 场景:正常关闭流程
- **当** 调用 `Runtime.Shutdown()`
- **那么** 并行卸载所有 Service 和 Worker，每个输出 stopping/stopped 日志，最后输出 all-service-closed、all-worker-closed、discovery-disconnected

### 需求:Service/Worker 生命周期日志
框架必须为 Service 和 Worker 的安装、启动、停止输出日志：
- `service-starting` / `service-started` (Info/Success): content 包含 name、id
- `service-stopping` / `service-stopped` (Info/Success): content 包含 name、id、reason
- `worker-starting` / `worker-started` (Info/Success): content 包含 name、id
- `worker-stopping` / `worker-stopped` (Info/Success): content 包含 name、id、reason
- `worker-on-error` (Error): Worker 启动失败时，content 包含 error

#### 场景:Service 正常启动
- **当** `Runtime.InstallService()` 被调用且启动成功
- **那么** 先输出 service-starting (Info)，完成后输出 service-started (Success)

#### 场景:Worker 启动失败
- **当** `Runtime.InstallWorker()` 被调用且启动失败
- **那么** 输出 install-worker-start (Error) 日志，content 包含 error 字段

#### 场景:Service 停止
- **当** `Runtime.UninstallService()` 被调用
- **那么** 输出 service-stopping (Info)，停止完成后输出 service-stopped (Success)

### 需求:Listener 管理日志
Service 管理 Listener 时必须输出日志：
- `install-listener` (Info): content 包含 name、id、meta（endpoint 信息）
- `listener-started` (Success): content 包含 name、id、meta
- `uninstall-listener` (Info): content 包含 name、id、meta
- `listener-stopped` (Success): content 包含 name、id、meta
- `listener-err` (Info): Listener 进入错误状态

#### 场景:安装 Listener
- **当** `BaseService.InstallListener()` 被调用且成功
- **那么** 先输出 install-listener (Info)，完成后输出 listener-started (Success)

#### 场景:卸载 Listener
- **当** Listener 被卸载
- **那么** 输出 uninstall-listener (Info)，完成后输出 listener-stopped (Success)

### 需求:RPC Connector 日志
RPC Connector 必须在以下事件输出日志：
- `connector-error` (Error): 连接错误，content 包含 error
- `connector-response-not-enabled` (Warn): 收到请求但 callback 未注册
- `parse-body-failed` (Warn): 数据包解析失败
- `handle-command-error` (Error): 命令处理失败
- `opcode-not-support` (Error): 不支持的 opcode
- `handle-notify-error` (Error): Notify 处理失败
- `event-handle-data` (Error): Request 处理失败

#### 场景:Connector 连接错误
- **当** Connection 发生错误
- **那么** 输出 Error 级别日志，event 为 `connector-error`

#### 场景:收到无法解析的数据包
- **当** 收到的数据包格式不正确
- **那么** 输出 Warn 级别日志，event 为 `parse-body-failed`

### 需求:Provider Sender 日志
Provider 管理 Sender 时必须输出日志：
- `sender-created` (Success): content 包含 id、listener（protocol、endpoint）、targetId、name
- `remove-sender` (Info): content 包含 name、id、listenerId、targetId
- `connector-off` (Error): sender 关闭时出错

#### 场景:创建 Sender
- **当** Provider 发现匹配的 endpoint 并创建新的 RPCSender
- **那么** 输出 Success 级别日志，event 为 `sender-created`

#### 场景:移除 Sender
- **当** endpoint 不再匹配或 Provider 关闭
- **那么** 输出 Info 级别日志，event 为 `remove-sender`

### 需求:Worker 组件和 Provider 管理日志
Worker 连接 Component 和注册 Provider 时必须输出日志：
- `register-provider` / `provider-started` (Info): content 包含 id、name、provider
- `unregister-provider` / `provider-unregistered` (Info): content 包含 id、name、provider
- `connect-component` / `component-connected` (Info): content 包含 id、name、component、version
- `disconnect-component` / `component-disconnected` (Info): content 包含 id、name、component

#### 场景:注册 Provider
- **当** `BaseWorker.RegisterProvider()` 被调用且成功
- **那么** 输出 register-provider (Info)，完成后输出 provider-started (Info)

#### 场景:连接 Component
- **当** `BaseWorker.ConnectComponent()` 被调用且成功
- **那么** 输出 connect-component (Info)，完成后输出 component-connected (Info)

### 需求:Discovery 注册日志
Discovery 注册/注销操作必须输出日志：
- `discovery-register-service` (Error): 注册服务失败
- `discovery-unregister-service` (Error): 注销服务失败
- `discovery-register-worker` (Error): 注册 Worker 失败
- `discovery-unregister-worker` (Error): 注销 Worker 失败

#### 场景:服务注册失败
- **当** Discovery 注册服务操作返回错误
- **那么** 输出 Error 级别日志，event 为 `discovery-register-service`，content 包含 error、name、id

### 需求:Goroutine Panic 恢复日志
框架启动的所有 goroutine 必须在入口处添加 `defer recover()` 机制。当 panic 被捕获时，必须输出日志：
- `goroutine-panic` (Error): content 包含 recover 值

#### 场景:Worker goroutine panic
- **当** Worker 启动的 goroutine 发生 panic
- **那么** panic 被捕获，输出 Error 级别日志，event 为 `goroutine-panic`，content 包含 recover 值

#### 场景:RPC 读写 goroutine panic
- **当** RPC 层的 goroutine 发生 panic
- **那么** panic 被捕获，输出 Error 级别日志，进程不会因 panic 而崩溃

### 需求:日志 Category 命名规范
所有框架日志必须使用以下 category 命名：
- `runtime`: Runtime 生命周期事件
- `process`: 进程信号处理
- `connector`: RPC Connector 连接事件
- `listener`: RPC Listener 事件
- `provider.<name>`: Provider 层事件（name 为服务名）
- `discovery`: Discovery 注册/注销事件

#### 场景:验证 category 命名
- **当** 框架输出任何日志
- **那么** category 必须是上述预定义值之一
