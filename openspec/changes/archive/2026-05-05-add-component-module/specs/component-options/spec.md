## ADDED Requirements

### Requirement: LoadOptions delegates to impl
baseComponent 的 LoadOptions 方法 SHALL 调用 impl.SetOptions(opts)。

#### Scenario: LoadOptions calls SetOptions
- **WHEN** 调用 baseComponent.LoadOptions(opts)
- **THEN** SHALL 将 opts 原样传递给 impl.SetOptions(opts) 并返回其结果

#### Scenario: LoadOptions with nil opts
- **WHEN** 调用 baseComponent.LoadOptions(nil) 且 impl.SetOptions(nil) 返回 nil
- **THEN** SHALL 返回 nil

### Requirement: SetOptions direct assignment
具体组件的 SetOptions 实现 SHALL 将传入的 opts 直接赋值给内部 options 字段，不进行 merge 或 diff。

#### Scenario: SetOptions overwrites previous options
- **WHEN** 具体组件的 options 为 {A: 1}，调用 SetOptions({B: 2})
- **THEN** 内部 options SHALL 变为 {B: 2}，先前的 {A: 1} SHALL 被完全替换

### Requirement: LoadOptions before Start
LoadOptions SHALL 在 Start 之前调用。调用顺序由消费者保证。

#### Scenario: Options applied before connect
- **WHEN** 先调用 LoadOptions(opts) 再调用 Start(ctx)
- **THEN** impl.Connect(ctx) 执行时 SHALL 能读取到通过 SetOptions 设置的 options
