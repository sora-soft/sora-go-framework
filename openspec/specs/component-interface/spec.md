### Requirement: Component interface definition
系统 SHALL 定义 `Component` 接口，包含以下方法：
- `Start(ctx context.Context) error`
- `Stop() error`
- `LoadOptions(opts any) error`
- `GetMetaInfo() ComponentMetadata`

#### Scenario: Component interface satisfies external consumers
- **WHEN** 消费者持有一个 `Component` 接口实例
- **THEN** 消费者可以调用 Start、Stop、LoadOptions、GetMetaInfo 四个方法

### Requirement: Internal implementation interface
系统 SHALL 定义内部实现接口（无导出名），具体组件 MUST 实现以下方法：
- `Connect(ctx context.Context) error`
- `Disconnect() error`
- `SetOptions(opts any) error`
- `GetOptions() any`

#### Scenario: Concrete component implements internal interface
- **WHEN** 创建一个具体组件（如 EtcdComponent）
- **THEN** 该组件 MUST 实现 Connect、Disconnect、SetOptions、GetOptions 四个方法

### Requirement: ComponentMetadata type
系统 SHALL 定义 `ComponentMetadata` 结构体，包含以下字段：
- `Name string`
- `Ready bool`
- `Version string`
- `Options any`

#### Scenario: ComponentMetadata serialization
- **WHEN** 调用 GetMetaInfo 获取 ComponentMetadata
- **THEN** 返回的结构体 SHALL 包含 name、ready、version、options 四个字段
