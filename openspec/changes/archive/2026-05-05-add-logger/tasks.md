## 1. Core Types

- [x] 1.1 Create `pkg/logger/` package directory and `logger.go` file: define `LogLevel` type with constants (Debug=1, Info=2, Success=3, Warn=4, Error=5, Fatal=6) and `String()` method
- [x] 1.2 Define `LoggerData` struct with fields: Time, Identify, Category, Level, Error, Content, Position, Stack, PID
- [x] 1.3 Define `LoggerOutput` interface with `Log(data LoggerData)` and `Close() error` methods

## 2. Logger Implementation

- [x] 2.1 Implement `NewLogger(identify string) *Logger` constructor
- [x] 2.2 Implement `AddOutput(output LoggerOutput) *Logger` method (returns *Logger for chaining)
- [x] 2.3 Implement private `write(level, category, error, content)` method: construct LoggerData with time.Now(), os.Getpid(), runtime.Caller() for position, json.Marshal for content, conditional runtime.Stack() for Error/Fatal, dispatch to all outputs
- [x] 2.4 Implement public log methods: Debug, Info, Success, Warn, Error, Fatal — each calling write() with appropriate parameters
- [x] 2.5 Implement `Error()` method with exerror level mapping using errors.As
- [x] 2.6 Implement `Close() error` method calling Close() on all outputs

## 3. ConsoleOutput Implementation

- [x] 3.1 Create `pkg/logger/console_output.go`: define `ConsoleOutput` struct with levels map and colors map
- [x] 3.2 Implement `NewConsoleOutput(levels ...LogLevel) *ConsoleOutput` with default ANSI color map
- [x] 3.3 Implement `WithColors(colors map[LogLevel]string)` option for custom color overrides
- [x] 3.4 Implement `Log(data LoggerData)` with level whitelist filtering, ANSI color wrapping, and comma-separated format output (timeString,LEVEL,identify,category,position,content)
- [x] 3.5 Implement `Close() error` as no-op returning nil
