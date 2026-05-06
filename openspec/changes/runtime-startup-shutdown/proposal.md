## Why

Runtime 目前是纯被动的注册表，Startup 和 Shutdown 均为空实现。框架缺少统一的启动/关闭编排入口——消费代码需要手动编排 backend 连接、node 注册、service/worker 停止等步骤。将 Runtime 升级为框架级别的生命周期管理者，提供 `Startup(ctx, node, backend)` 和 `Shutdown()` 方法，实现一键启动和优雅关闭。

## What Changes

- **BREAKING**: `Runtime.Startup()` 签名从 `Startup() error` 变更为 `Startup(ctx context.Context, node *NodeRunner, backend discovery.Backend) error`
- **BREAKING**: `Runtime` 从 `pkg/runner/runtime` 包合并到 `pkg/runner` 包，所有 `runtime.RT` 引用变为 `RT`（同包内）或 `runner.RT`（外部包）
- 删除 `pkg/runner/runtime/` 子包
- `Startup(ctx, node, backend)` 连接 discovery backend 并存储 node/backend 引用
- `Shutdown()` 并发停止所有 worker 和 service（不含 node），等待全部退出后停止 node，最后断开 discovery backend
- 新增 `GetNode()`、`GetBackend()`、`GetDiscovery()` 访问方法
- 更新 `node.go`、`woker.go`、`cmd/sora-test/main.go` 的 import 路径

## Capabilities

### New Capabilities
- `runtime-startup`: Runtime 接受 NodeRunner 和 discovery.Backend 参数执行启动，连接 backend 并存储引用
- `runtime-shutdown`: Runtime 执行优雅关闭——并发停止所有 worker/service，等待退出后停止 node，最后断开 backend

### Modified Capabilities
- `runtime-component-registry`: Runtime struct 新增 node、backend 字段及对应访问方法
- `runtime-worker-service-registry`: Shutdown 新增停止所有已注册 service/worker 的行为
- `runtime-process-metadata`: nodeId 的读写与新字段的并发安全统一管理
- `node-runner`: NodeRunner.Startup 中 `runtime.RT` 引用方式变更（同包直接访问）

## Impact

- **包结构**: `pkg/runner/runtime/` 目录整体删除，`runtime.go` 移至 `pkg/runner/`
- **Import 路径**: `pkg/runner/node.go`、`pkg/runner/woker.go` 删除对 `pkg/runner/runtime` 的 import；`cmd/sora-test/main.go` 改为 import `pkg/runner`
- **API 签名**: `Startup` 签名变更，任何外部调用需适配
- **依赖**: Runtime 新增对 `pkg/discovery` 的 import
