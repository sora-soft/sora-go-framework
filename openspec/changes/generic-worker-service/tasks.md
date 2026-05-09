## 1. 泛型化 BaseWorker

- [x] 1.1 将 `baseWorker` struct 改写为 `BaseWorker[R Runner]`，命名字段 `runner R` 替代 `types.Runner` 嵌入
- [x] 1.2 实现 `NewWorker[R Runner](name string, runner R, opts WorkerOptions) *BaseWorker[R]` 构造函数，含 WorkerAware 注入
- [x] 1.3 适配 `BaseWorker[R]` 的所有方法：Start、Stop、Go、GetId、GetMetadata、ConnectComponent、RegisterProvider、disconnectComponents、stopProviders
- [x] 1.4 添加 `Runner() R` 方法

## 2. 泛型化 BaseService

- [x] 2.1 将 `baseService` struct 改写为 `BaseService[R Runner]`，嵌入 `*BaseWorker[R]` 替代 `*baseWorker`
- [x] 2.2 实现 `NewService[R Runner](name string, runner R, opts ServiceOptions) *BaseService[R]` 构造函数，含 ServiceAware/WorkerAware 注入
- [x] 2.3 适配 `BaseService[R]` 的所有方法：Start、Stop、InstallListener、stopListeners、ListenLifeCycle、GetMetadata
- [x] 2.4 添加 `Runner() R` 方法

## 3. 适配 NodeRunner

- [x] 3.1 更新 `NodeRunner` 构造逻辑，使用 `NewService[*NodeRunner]` 泛型构造函数
- [x] 3.2 验证 NodeRunner.Startup 通过 svc.InstallListener 安装 listeners 的流程正常
- [x] 3.3 验证 NodeRunner.StateData 通过 svc.GetMetadata 获取元数据的流程正常

## 4. 验证

- [x] 4.1 编译通过，无类型错误
- [x] 4.2 `BaseWorker[R]` 满足 `types.Worker` 接口
- [x] 4.3 `BaseService[R]` 满足 `types.Service` 接口
- [x] 4.4 Runtime 注册/管理逻辑无需修改，兼容新类型
