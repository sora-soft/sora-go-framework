## ADDED Requirements

### 需求:EtcdClientConfig 独立类型
系统必须定义独立的 `EtcdClientConfig` 结构体，包含 `Endpoints`（[]string）、`DialTimeout`（time.Duration）、`Username`（string，可选）、`Password`（string，可选）字段，用于承载 etcd 客户端连接配置。

#### 场景:EtcdClientConfig 字段完整
- **当** 查看代码
- **那么** `EtcdClientConfig` 作为独立导出类型存在，包含 Endpoints、DialTimeout、Username、Password 字段

### 需求:EtcdComponentOptions 嵌套结构
系统必须定义 `EtcdComponentOptions` 结构体，包含嵌套字段 `Etcd`（EtcdClientConfig 类型）、`TTL`（int64，单位秒）、`Prefix`（string），替代原有的平铺 `EtcdOptions`。

#### 场景:选项结构嵌套
- **当** 查看代码
- **那么** `EtcdComponentOptions` 包含 `Etcd EtcdClientConfig`、`TTL int64`、`Prefix string` 三个字段

### 需求:EtcdComponentOptions 必填字段校验
`SetOptions` 必须对 `EtcdComponentOptions` 的必填字段进行校验：`Etcd.Endpoints` 不得为空、`Etcd.DialTimeout` 不得为零值、`TTL` 必须 > 0、`Prefix` 不得为空字符串。校验失败必须返回描述性错误。

#### 场景:Endpoints 为空
- **当** 调用 `SetOptions` 且 `Etcd.Endpoints` 为空切片
- **那么** 返回错误，描述 Endpoints 不得为空

#### 场景:DialTimeout 为零
- **当** 调用 `SetOptions` 且 `Etcd.DialTimeout` 为 0
- **那么** 返回错误，描述 DialTimeout 不得为零

#### 场景:TTL 为零或负数
- **当** 调用 `SetOptions` 且 `TTL <= 0`
- **那么** 返回错误，描述 TTL 必须 > 0

#### 场景:Prefix 为空
- **当** 调用 `SetOptions` 且 `Prefix` 为空字符串
- **那么** 返回错误，描述 Prefix 不得为空

#### 场景:所有字段有效
- **当** 调用 `SetOptions` 且所有必填字段有效
- **那么** 选项被接受，不返回错误

### 需求:Connect 适配嵌套结构
`Connect` 方法必须从 `options.Etcd` 读取客户端配置（Endpoints、DialTimeout、Username、Password），而非从 options 顶层直接读取。

#### 场景:Connect 使用嵌套配置
- **当** 调用 `Connect` 且 options 已正确设置
- **那么** 使用 `options.Etcd.Endpoints`、`options.Etcd.DialTimeout` 等字段创建 etcd 客户端

### 需求:GetOptions 返回 EtcdComponentOptions
`GetOptions` 必须返回 `EtcdComponentOptions` 值（非指针），类型与 `SetOptions` 输入一致。

#### 场景:GetOptions 返回完整选项
- **当** 调用 `GetOptions`
- **那么** 返回 `EtcdComponentOptions` 类型的值，包含完整的嵌套结构

### 需求:EtcdOptions 类型移除
系统必须移除原有的 `EtcdOptions` 类型，所有引用必须改为 `EtcdComponentOptions`。

#### 场景:EtcdOptions 不存在
- **当** 查看代码
- **那么** `EtcdOptions` 类型不再存在，所有使用处已替换为 `EtcdComponentOptions`

### 需求:Username 和 Password 可选
`EtcdClientConfig` 的 `Username` 和 `Password` 字段必须为可选字段。当 `Username` 为空时，`Connect` 不得设置客户端的用户名和密码。

#### 场景:无认证信息
- **当** 调用 `SetOptions` 且 `Username` 为空
- **那么** `SetOptions` 不报错，`Connect` 创建客户端时不设置认证信息

#### 场景:有认证信息
- **当** 调用 `SetOptions` 且 `Username` 非空
- **那么** `Connect` 创建客户端时设置 Username 和 Password
