## 1. watchLoop Stopping 状态处理

- [x] 1.1 在 `provider.go` 的 `watchLoop` 中，对已存在的 endpoint 增加 State == Stopping 检查，若命中则调用 `removeSenderLocked`
- [x] 1.2 在 `provider.go` 的 `watchLoop` 中，对新发现的 endpoint（不在 currentIds 中）增加 State == Stopping 检查，若命中则跳过 `addSenderLocked`

## 2. 验证

- [x] 2.1 确认 `endpointChanged` 函数无需修改（State 变化走新分支，不走 recreate 路径）
- [x] 2.2 运行项目构建确认编译通过
