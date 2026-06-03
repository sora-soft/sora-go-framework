## ADDED Requirements

### 需求:BaseWorker 泛型化
BaseWorker 必须为泛型结构体 `BaseWorker[R Runner]`，通过命名字段 `runner R` 持有具体 Runner 类型，并满足 `types.Worker` 接口。

#### 场景:创建泛型 Worker
- **当** 调用 `NewWorker[R](name, runner, opts)` 且 R 为具体 Runner 类型
- **那么** 返回的 `*BaseWorker[R]` 满足 `types.Worker` 接口，且 `.Runner()` 返回原始 R 类型

#### 场景:Worker 启动委托
- **当** 调用 `BaseWorker[R].Start(ctx)`
- **那么** 内部调用 `runner.Startup(ctx)` 启动具体 Runner

#### 场景:Worker 停止清理
- **当** 调用 `BaseWorker[R].Stop()`
- **那么** 依次执行：设置 stopping 状态 → cancel context → 等待 goroutine → runner.Shutdown() → 停止 providers → 断开 components → 设置 stopped 状态

### 需求:BaseService 泛型化
BaseService 必须为泛型结构体 `BaseService[R Runner]`，嵌入 `*BaseWorker[R]`，管理 listeners 注册表，并满足 `types.Service` 接口。

#### 场景:创建泛型 Service
- **当** 调用 `NewService[R](name, runner, opts)` 且 R 为具体 Runner 类型
- **那么** 返回的 `BaseService[R]` 满足 `types.Service` 接口，且 `.Runner()` 返回原始 R 类型

#### 场景:Service 启动
- **当** 调用 `BaseService[R].Start(ctx)`
- **那么** 直接委托给嵌入的 `*BaseWorker[R].Start(ctx)`，即 `runner.Startup(ctx)`

#### 场景:Service 停止
- **当** 调用 `BaseService[R].Stop()`
- **那么** 依次执行：停止 listeners → cancel context → 等待 goroutine → runner.Shutdown() → 停止 providers → 断开 components → 设置 stopped 状态

### 需求:InstallListener 注册表
BaseService 必须维护已安装 listener 的注册表，`InstallListener` 启动 listener、注册到 discovery 并追踪生命周期。

#### 场景:安装 Listener
- **当** Runner 在 Startup 中调用 `svc.InstallListener(ctx, l)`
- **那么** listener 启动成功后追加到 `BaseService.listeners`，自动注册 endpoint 到 discovery

#### 场景:Listener 状态监听
- **当** listener 进入 Ready 或 Stopping 状态
- **那么** 自动注册 endpoint 到 discovery registry

#### 场景:Listener 停止或出错
- **当** listener 进入 Stopped 或 Error 状态
- **那么** 自动从 discovery 注销 endpoint，移除监听器

#### 场景:Service 停止时清理 Listeners
- **当** `BaseService[R].Stop()` 被调用
- **那么** 在 runner.Shutdown() 之前停止所有已注册的 listeners

### 需求:Aware 反向引用注入
构造函数必须检查 Runner 是否实现 `WorkerAware` 或 `ServiceAware`，并自动注入反向引用。

#### 场景:Runner 实现 ServiceAware
- **当** `NewService[R]` 构造时，R 实现了 `ServiceAware` 接口
- **那么** 构造函数自动调用 `runner.SetService(svc)`，svc 为新创建的 `BaseService[R]`

#### 场景:Runner 仅实现 WorkerAware
- **当** `NewService[R]` 构造时，R 未实现 `ServiceAware` 但实现了 `WorkerAware`
- **那么** 构造函数调用 `runner.SetWorker(svc)`，svc 作为 Worker 注入

#### 场景:NewWorker 中的 Aware 注入
- **当** `NewWorker[R]` 构造时，R 实现了 `WorkerAware`
- **那么** 构造函数自动调用 `runner.SetWorker(w)`，w 为新创建的 `BaseWorker[R]`

### 需求:Runner() 类型安全访问
BaseWorker 和 BaseService 必须提供 `Runner() R` 方法，返回持有的具体 Runner 类型。

#### 场景:通过 BaseWorker 获取 Runner
- **当** 持有 `*BaseWorker[*NodeRunner]` 并调用 `.Runner()`
- **那么** 返回 `*NodeRunner` 类型，无需类型断言

#### 场景:通过 BaseService 获取 Runner
- **当** 持有 `*BaseService[*NodeRunner]` 并调用 `.Runner()`
- **那么** 返回 `*NodeRunner` 类型，无需类型断言
