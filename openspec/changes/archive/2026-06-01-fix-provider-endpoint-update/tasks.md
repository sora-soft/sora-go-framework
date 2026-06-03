## 1. 变更检测函数

- [x] 1.1 在 `pkg/rpc/provider/provider.go` 中新增 `endpointChanged(a, b discovery.EndpointMeta) bool` 函数，比较 Endpoint、Protocol、Codecs 三个字段

## 2. watchLoop 逻辑修改

- [x] 2.1 修改 `watchLoop` 中已存在 endpoint 的处理分支：当 `endpointChanged` 返回 true 时，调用 `removeSenderLocked` 销毁旧 sender，再调用 `addSenderLocked` 用新 endpoint 创建 sender

## 3. 验证

- [x] 3.1 编写 `endpointChanged` 的单元测试，覆盖：地址变更、协议变更、Codecs 变更、非关键字段变更、完全相同等场景
- [x] 3.2 运行 `go build ./...` 确认编译通过
- [x] 3.3 运行 `go test ./pkg/rpc/provider/...` 确认测试通过
