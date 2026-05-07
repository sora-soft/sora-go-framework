## 1. 包合并

- [x] 1.1 将 `pkg/runner/runtime/runtime.go` 内容复制到 `pkg/runner/runtime.go`，修改 package 声明为 `package runner`
- [x] 1.2 在 `pkg/runner/runtime.go` 中添加 `pkg/discovery` import
- [x] 1.3 删除 `pkg/runner/runtime/` 目录

## 2. Runtime 结构体变更

- [x] 2.1 在 Runtime struct 中添加 `nodeMu sync.RWMutex`、`node *NodeRunner`、`backend discovery.Backend` 字段
- [x] 2.2 实现 `GetNode() *NodeRunner` 方法（读锁保护）
- [x] 2.3 实现 `GetBackend() discovery.Backend` 方法（读锁保护）
- [x] 2.4 实现 `GetDiscovery() discovery.Discovery` 方法（读锁保护，backend 为 nil 时返回 nil）

## 3. Startup 实现

- [x] 3.1 将 `Startup() error` 签名变更为 `Startup(ctx context.Context, node *NodeRunner, backend discovery.Backend) error`
- [x] 3.2 实现 Startup 方法：调用 `backend.Connect(ctx)`，成功后用写锁存储 node 和 backend 引用

## 4. Shutdown 实现

- [x] 4.1 将 `Shutdown() error` 变更为完整实现：并发停止所有 Service 和 Worker（排除 nodeId 对应的 service）
- [x] 4.2 实现并发停止逻辑：收集所有 service ID 和 worker ID，排除 nodeId 对应的 service，用 goroutine + WaitGroup 并发调用 UninstallService/UninstallWorker
- [x] 4.3 实现 node 停止阶段：等待并发阶段完成后，用 nodeId 调用 UninstallService 停止 node service
- [x] 4.4 实现 backend 断开阶段：用写锁取出并清除 backend 引用，调用 backend.Disconnect()
- [x] 4.5 实现错误收集：收集所有 Stop/Disconnect 错误，返回第一个错误

## 5. Import 更新

- [x] 5.1 更新 `pkg/runner/node.go`：删除 `pkg/runner/runtime` import，将 `runtime.RT` 替换为 `RT`
- [x] 5.2 更新 `pkg/runner/woker.go`：删除 `pkg/runner/runtime` import，将 `runtime.RT` 替换为 `RT`
- [x] 5.3 更新 `cmd/sora-test/main.go`：将 import 从 `pkg/runner/runtime` 改为 `pkg/runner`，将 `runtime.RT` 替换为 `runner.RT`

## 6. 验证

- [x] 6.1 确认编译通过（`go build ./...`）
- [x] 6.2 确认现有测试通过（`go test ./...`）
