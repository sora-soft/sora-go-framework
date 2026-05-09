## 1. 类型定义

- [x] 1.1 定义 `EtcdClientConfig` 结构体（Endpoints、DialTimeout、Username、Password），移除原有 `EtcdOptions` 类型
- [x] 1.2 定义 `EtcdComponentOptions` 结构体（Etcd EtcdClientConfig、TTL int64、Prefix string）

## 2. 校验逻辑

- [x] 2.1 在 `SetOptions` 中添加字段校验：Etcd.Endpoints 非空、Etcd.DialTimeout > 0、TTL > 0、Prefix 非空
- [x] 2.2 定义校验错误变量（ErrEndpointsEmpty、ErrDialTimeoutZero、ErrTTLInvalid、ErrPrefixEmpty）

## 3. 连接逻辑适配

- [x] 3.1 修改 `Connect` 方法，从 `options.Etcd` 读取客户端配置
- [x] 3.2 修改 `EtcdComponent` 的 `options` 字段类型为 `*EtcdComponentOptions`

## 4. 测试更新

- [x] 4.1 更新所有测试用例，使用新的 `EtcdComponentOptions` 嵌套结构
- [x] 4.2 添加校验相关测试：Endpoints 为空、DialTimeout 为零、TTL 无效、Prefix 为空
- [x] 4.3 更新 `GetOptions` 返回类型相关测试
