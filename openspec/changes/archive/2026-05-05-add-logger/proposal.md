## Why

项目当前没有任何日志基础设施，`pkg/` 下的库代码完全静默——不打印、不记录、不输出任何诊断信息。调试、监控、问题排查都缺乏结构化的可观测性手段。引入统一的 logger 包为所有组件提供一致的日志入口和可插拔的输出策略。

## What Changes

- 新增 `pkg/logger` 包，提供 `Logger` 结构体作为日志入口，支持 Debug/Info/Success/Warn/Error/Fatal 六个日志等级
- 新增 `LoggerOutput` 接口，定义日志输出的抽象契约（`Log` + `Close`）
- 新增 `ConsoleOutput`，作为内置的控制台输出实现，支持按等级配置 ANSI 颜色和等级白名单过滤
- `Logger.Error()` 方法内置 `exerror.ExError` 等级映射：Fatal→Fatal，Unexpected→Error，Expected→Warn，Silent→不输出
- 只有 Error/Fatal 级别才捕获 `runtime.Stack()`
- 日志过滤责任完全在 Output 层，每个 Output 通过等级白名单决定是否输出

## Capabilities

### New Capabilities
- `logger-core`: Logger 结构体、LogLevel 枚举、LoggerData 数据结构、LoggerOutput 接口、AddOutput/Close 生命周期管理、exerror 等级映射逻辑
- `console-output`: ConsoleOutput 实现，等级白名单过滤，可配置 ANSI 颜色，逗号分隔格式化输出

### Modified Capabilities

## Impact

- 新增 `pkg/logger/` 包（logger.go, output.go, console_output.go）
- Logger 依赖 `pkg/utility/exerror` 包（单向依赖，无循环风险）
- 无外部依赖引入（ANSI 颜色使用原生 escape code）
- 不影响现有代码，后续各组件可按需引入 Logger 实例
