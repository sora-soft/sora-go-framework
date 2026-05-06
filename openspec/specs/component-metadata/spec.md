### Requirement: GetMetaInfo returns ComponentMetadata
baseComponent SHALL 实现 GetMetaInfo 方法，返回 ComponentMetadata，其中：
- Name 取自 baseComponent.Name
- Ready 取自 baseComponent.ready
- Version 取自 baseComponent.Version
- Options 取自 impl.GetOptions() 的返回值

#### Scenario: GetMetaInfo for running component
- **WHEN** baseComponent 的 Name 为 "etcd-main"，Version 为 "0.0.0"，ready 为 true，impl.GetOptions() 返回 {endpoints: ["localhost:2379"]}
- **THEN** GetMetaInfo() SHALL 返回 ComponentMetadata{Name: "etcd-main", Ready: true, Version: "0.0.0", Options: {endpoints: ["localhost:2379"]}}

#### Scenario: GetMetaInfo for stopped component
- **WHEN** baseComponent 的 ready 为 false，impl.GetOptions() 返回 nil
- **THEN** GetMetaInfo() SHALL 返回 ComponentMetadata{Ready: false, Options: nil}

### Requirement: GetMetaInfo on Component interface
GetMetaInfo SHALL 是 Component 接口的方法，消费者通过 Component 接口调用。

#### Scenario: Consumer calls GetMetaInfo through interface
- **WHEN** 消费者持有 Component 接口类型的变量
- **THEN** 消费者 SHALL 能直接调用 GetMetaInfo() 获取 ComponentMetadata

### Requirement: GetOptions from impl
具体组件 SHALL 实现 GetOptions 方法，返回当前 options 的拷贝或引用。

#### Scenario: GetOptions returns current options
- **WHEN** 具体组件的内部 options 为 EtcdOptions{Endpoints: []string{"localhost:2379"}}
- **THEN** GetOptions() SHALL 返回该 options 值
