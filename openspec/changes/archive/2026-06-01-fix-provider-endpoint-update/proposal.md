## Why

Provider 的 watchLoop 在收到 discovery 快照时仅处理新增和移除的 endpoint，不检测已有 endpoint 的地址变更。当 listener 尚未就绪时注册了不完整地址（如端口为 0），后续 listener 更新了正确地址，已创建的 RpcSender 仍使用旧地址重试连接，导致永远无法连接成功。

## What Changes

- Provider watchLoop 增加对已有 endpoint 的关键字段变更检测（Endpoint、Protocol、Codecs）
- 当检测到变更时，销毁旧 RpcSender 并用新 endpoint 数据创建替换

## Capabilities

### New Capabilities

- `provider-endpoint-update`: Provider 对 discovery 快照中已有 endpoint 的地址变更进行检测并热替换 RpcSender

### Modified Capabilities

_(无规范级行为变更，仅修改 provider-lifecycle 中 watchLoop 的实现行为)_

## Impact

- `pkg/rpc/provider/provider.go` — watchLoop 逻辑变更，新增 endpoint 比较与 sender 热替换
- `pkg/rpc/provider/rpc_sender.go` — 无结构性变更，endpoint 字段已可同包访问
