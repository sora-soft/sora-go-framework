## 1. 结构体修改

- [x] 1.1 在 EtcdBackend 结构体中新增 `comp component.Component` 字段

## 2. Connect 方法修复

- [x] 2.1 在 Connect 中获取 component 后、类型断言前，调用 `c.Start(ctx)` 确保组件已连接
- [x] 2.2 将 Component 引用保存到 `b.comp` 字段

## 3. Disconnect 方法修复

- [x] 3.1 在 Disconnect 末尾添加 `b.comp.Stop()` 调用，释放组件引用计数
- [x] 3.2 确保 Stop 顺序：先清理 keepAliveCancel → watchers → lease → comp.Stop()

## 4. 清理 debug 代码

- [x] 4.1 移除 initFromEtcd 中的 println 调试语句（line 113, 116, 124）

## 5. 测试验证

- [x] 5.1 运行现有 etcd 测试确认通过（测试中手动 Start + Backend 内部 Start 应通过引用计数安全共存）
- [x] 5.2 确认 TestConnect_ComponentNotFound 和 TestConnect_ComponentNotEtcd 仍正确返回错误
