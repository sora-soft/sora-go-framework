## 1. Runtime 层日志

- [x] 1.1 在 `pkg/runtime/runtime.go` 的 `LoadConfig` 方法中添加 `load-config` Info 日志
- [x] 1.2 在 `pkg/runtime/runtime.go` 的 `Startup` 方法中添加 discovery 连接日志（`connect-discovery` 成功 Info / 失败 Fatal）、node 安装日志（`install-node` 成功 Info / 失败 Fatal）、`start-runtime-success` Success 日志
- [x] 1.3 在 `pkg/runtime/runtime.go` 的 `Startup` 方法中注册 SIGINT/SIGTERM 信号处理，输出 `process-command` Info 日志后调用 Shutdown
- [x] 1.4 在 `pkg/runtime/runtime.go` 的 `Shutdown` 方法中添加 `all-service-closed`、`all-worker-closed`、`discovery-disconnected` Info 日志
- [x] 1.5 在 `pkg/runtime/runtime.go` 的 `InstallService` 方法中添加 `service-starting` Info 日志和 `service-started` Success 日志（成功时），失败时添加 `install-service-start` Error 日志
- [x] 1.6 在 `pkg/runtime/runtime.go` 的 `InstallWorker` 方法中添加 `worker-starting` Info 日志和 `worker-started` Success 日志（成功时），失败时添加 `install-worker-start` Error 日志
- [x] 1.7 在 `pkg/runtime/runtime.go` 的 `UninstallService` 方法中添加 `service-stopping` Info 日志和 `service-stopped` Success 日志
- [x] 1.8 在 `pkg/runtime/runtime.go` 的 `UninstallWorker` 方法中添加 `worker-stopping` Info 日志和 `worker-stopped` Success 日志

## 2. Worker/Service 层日志

- [x] 2.1 在 `pkg/runner/woker.go` 的 `Start` 方法中添加 `worker-on-error` Error 日志（启动失败时）
- [x] 2.2 在 `pkg/runner/woker.go` 的 `ConnectComponent` 方法中添加 `connect-component` / `component-connected` Info 日志
- [x] 2.3 在 `pkg/runner/woker.go` 的 `RegisterProvider` 方法中添加 `register-provider` / `provider-started` Info 日志
- [x] 2.4 在 `pkg/runner/woker.go` 的 `disconnectComponents` 和 `stopProviders` 方法中添加 `disconnect-component` / `component-disconnected`、`unregister-provider` / `provider-unregistered` Info 日志
- [x] 2.5 在 `pkg/runner/service.go` 的 `InstallListener` 方法中添加 `install-listener` / `listener-started` Info/Success 日志
- [x] 2.6 在 `pkg/runner/service.go` 的 `Stop` 方法中添加 `uninstall-listener` / `listener-stopped` Info/Success 日志

## 3. RPC Connector 层日志

- [x] 3.1 在 `pkg/rpc/connector.go` 的连接错误处添加 `connector-error` Error 日志
- [x] 3.2 在 `pkg/rpc/connector.go` 的 `handlePacket` 方法中添加 `opcode-not-support` Error 日志和 `parse-body-failed` Warn 日志
- [x] 3.3 在 `pkg/rpc/connector.go` 的 `handleCommand` 方法中添加 `handle-command-error` Error 日志

## 4. RPC Listener 层日志

- [x] 4.1 在 `pkg/rpc/listener.go` 的 `newConnector` 方法中添加会话打开/关闭日志
- [x] 4.2 在 `pkg/rpc/listener.go` 的 `Start`/`Stop` 方法中添加 listener 启停日志

## 5. Provider 层日志

- [x] 5.1 在 `pkg/rpc/provider/provider.go` 的 `addSenderLocked` 方法中添加 `sender-created` Success 日志
- [x] 5.2 在 `pkg/rpc/provider/provider.go` 的 `removeSenderLocked` 方法中添加 `remove-sender` Info 日志
- [x] 5.3 在 `pkg/rpc/provider/rpc_sender.go` 的 `connectLoop` 方法中添加连接成功/失败日志
- [x] 5.4 在 `pkg/rpc/provider/rpc_sender.go` 的 `Destroy` 方法中添加 `connector-off` Error 日志（关闭失败时）

## 6. Discovery 注册日志

- [x] 6.1 在 `pkg/runtime/runtime.go` 的 InstallService/InstallWorker 生命周期监听 goroutine 中添加 `discovery-register-service`/`discovery-register-worker` Error 日志（注册失败时）和 `discovery-unregister-service`/`discovery-unregister-worker` Error 日志（注销失败时）

## 7. Goroutine Panic 恢复

- [x] 7.1 在 `pkg/runtime/runtime.go` 中所有框架启动的 goroutine 入口添加 `defer recover()` + `goroutine-panic` Error 日志
- [x] 7.2 在 `pkg/rpc/listener.go` 的 `acceptLoop` 和 `newConnector` goroutine 入口添加 `defer recover()` + 日志
- [x] 7.3 在 `pkg/rpc/provider/rpc_sender.go` 的 `connectLoop` goroutine 入口添加 `defer recover()` + 日志
- [x] 7.4 在 `pkg/rpc/connector.go` 的 `readLoop` goroutine 入口添加 `defer recover()` + 日志
