## Context

项目 `pkg/` 下有三个核心包：`rpc`（RPC 框架）、`runner`（生命周期管理）、`utility`（工具集含 exerror）。目前所有库代码完全静默，没有任何日志输出能力。`exerror` 包已建立错误等级体系（Fatal/Unexpected/Expected/Silent），但没有消费方对其进行分级处理。

## Goals / Non-Goals

**Goals:**
- 提供统一的 `Logger` 结构体作为所有组件的日志入口
- 通过 `LoggerOutput` 接口实现输出策略的可插拔
- 内置 `exerror` 等级到 `LogLevel` 的映射，让 `Logger.Error()` 能智能分发
- 提供开箱即用的 `ConsoleOutput` 实现
- 保持最小依赖（零外部库引入）

**Non-Goals:**
- 不引入全局 logger 实例
- 不支持 `context.Context` 参与（未来可扩展）
- 不提供 FileOutput 等其他 Output 实现（后续 change 中添加）
- 不做日志异步缓冲或批量写入

## Decisions

### 1. Logger 和 Output 分层

Logger 负责日志数据构造（时间、调用位置、堆栈、序列化），Output 负责格式化和输出。Logger 无状态地遍历所有 Output 并调用 `Log()`。

**替代方案：** Logger 内置格式化逻辑直接输出。被否决——无法支持多种输出目标。

### 2. 过滤责任在 Output 层

每个 Output 持有一个 `map[LogLevel]struct{}` 白名单。`Log()` 被调用时先检查 level 是否在白名单中，不在则直接 return。Logger 层不做任何过滤。

**替代方案：** Logger 持有 minLevel 阈值。被否决——阈值只能做 >= 过滤，无法支持非连续的等级选择。

### 3. exerror 等级映射内置在 Logger.Error()

`Error()` 方法通过 `errors.As` 提取 `*exerror.ExError`，按其 Level 字段映射：
- `LevelFatal` → `LogLevelFatal`
- `LevelUnexpected`（含默认） → `LogLevelError`
- `LevelExpected` → `LogLevelWarn`
- `LevelSilent` → 不输出

普通 `error` 统一映射到 `LogLevelError`。

**替代方案：** Logger 不感知 exerror，由调用方判断。被否决——增加调用方负担，且与 TS 参考设计一致。

### 4. 只有 Error/Fatal 捕获堆栈

`runtime.Stack()` 有性能开销。只在 Error 和 Fatal 级别的 `write()` 调用中执行堆栈捕获，其他级别 `LoggerData.Stack` 为 nil。

### 5. Content 序列化只用 json.Marshal

不做 fmt.Sprint fallback。如果 Marshal 失败，Content 设为空字符串。

### 6. ConsoleOutput 使用原生 ANSI escape code

不引入第三方 color 库。通过 `map[LogLevel]string` 存储 ANSI code，支持用户自定义覆盖。

### 7. AddOutput 返回 *Logger

支持链式调用：`logger.AddOutput(a).AddOutput(b)`。

### 8. Logger.Close() 遍历所有 Output 调用 Close()

用于输出目标的优雅关闭（刷新缓冲区、关闭文件等）。ConsoleOutput 的 Close() 为空操作。

## Risks / Trade-offs

- **json.Marshal 失败** → Content 为空字符串，信息丢失。可接受，因为日志参数应使用可序列化类型。
- **Logger 不过滤** → 每条日志都会构造完整 LoggerData 并推给所有 Output，即使所有 Output 都会过滤掉。高频 Debug 日志场景下有一定性能开销。可接受，后续可通过在 Logger 层加缓存级别的快速路径优化。
- **ANSI 颜色在 Windows 终端显示不一致** → 不做特殊处理，现代 Windows Terminal 已支持 ANSI。
- **exerror 耦合** → Logger 依赖 exerror 包。可接受，方向单向且 exerror 零外部依赖。
