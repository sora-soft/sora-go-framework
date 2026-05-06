### Requirement: LogLevel enumeration
The system SHALL define a `LogLevel` type as `int` with six levels: Debug(1), Info(2), Success(3), Warn(4), Error(5), Fatal(6). Each level SHALL have a `String()` method returning its uppercase name.

#### Scenario: LogLevel values are sequential
- **WHEN** LogLevel constants are inspected
- **THEN** Debug=1, Info=2, Success=3, Warn=4, Error=5, Fatal=6

#### Scenario: LogLevel String returns uppercase name
- **WHEN** `LogLevelWarn.String()` is called
- **THEN** the result is "WARN"

### Requirement: LoggerData structure
The system SHALL define a `LoggerData` struct with fields: Time(time.Time), Identify(string), Category(string), Level(LogLevel), Error(error), Content(string), Position(string), Stack([]byte), PID(int).

#### Scenario: LoggerData captures all log entry metadata
- **WHEN** a LoggerData is constructed by Logger.write()
- **THEN** it contains the current time, logger identify, caller-provided category, resolved log level, optional error, serialized content string, caller file:line position, optional stack bytes, and current process ID

### Requirement: Logger construction
The system SHALL provide a `NewLogger(identify string) *Logger` constructor. The Logger SHALL store the identify string and an empty output list.

#### Scenario: Logger is created with identify
- **WHEN** `NewLogger("my-service")` is called
- **THEN** a Logger is returned with identify "my-service" and zero outputs

### Requirement: Logger output management
The system SHALL provide an `AddOutput(output LoggerOutput) *Logger` method that appends the output to the Logger's output list and returns the Logger for chaining. The system SHALL provide a `Close() error` method that calls `Close()` on all registered outputs and returns the first error encountered.

#### Scenario: AddOutput chains
- **WHEN** `logger.AddOutput(a).AddOutput(b)` is called
- **THEN** the Logger has two outputs [a, b]

#### Scenario: Close delegates to all outputs
- **WHEN** `logger.Close()` is called
- **THEN** `Close()` is called on every registered output

### Requirement: Logger log methods
The system SHALL provide methods: `Debug(category string, content any)`, `Info(category string, content any)`, `Success(category string, content any)`, `Warn(category string, content any)`, `Error(category string, err error, content any)`, `Fatal(category string, err error, content any)`.

#### Scenario: Debug logs with Debug level
- **WHEN** `logger.Debug("cat", data)` is called
- **THEN** LoggerData.Level is LogLevelDebug and LoggerData.Error is nil

#### Scenario: Fatal logs with Fatal level
- **WHEN** `logger.Fatal("cat", err, data)` is called
- **THEN** LoggerData.Level is LogLevelFatal and LoggerData.Error is err

### Requirement: Content serialization
The system SHALL serialize the `content` parameter using `json.Marshal`. If marshaling fails, Content SHALL be set to an empty string.

#### Scenario: Content is JSON serialized
- **WHEN** `logger.Info("cat", map[string]string{"key":"val"})` is called
- **THEN** LoggerData.Content is `{"key":"val"}`

#### Scenario: Content falls back to empty on marshal failure
- **WHEN** `json.Marshal(content)` returns an error
- **THEN** LoggerData.Content is ""

### Requirement: Caller position capture
The system SHALL capture the caller's file name and line number using `runtime.Caller` with appropriate skip depth, formatted as `"filename.go:lineNumber"`. The skip depth SHALL exclude the Logger's internal methods.

#### Scenario: Position reflects caller location
- **WHEN** `logger.Info("cat", nil)` is called from `connector.go` line 42
- **THEN** LoggerData.Position is "connector.go:42"

### Requirement: Stack capture for Error and Fatal only
The system SHALL capture `runtime.Stack()` ONLY when the log level is Error or Fatal. For all other levels, LoggerData.Stack SHALL be nil.

#### Scenario: Debug does not capture stack
- **WHEN** `logger.Debug("cat", nil)` is called
- **THEN** LoggerData.Stack is nil

#### Scenario: Error captures stack
- **WHEN** `logger.Error("cat", err, nil)` is called
- **THEN** LoggerData.Stack is non-nil

#### Scenario: Fatal captures stack
- **WHEN** `logger.Fatal("cat", err, nil)` is called
- **THEN** LoggerData.Stack is non-nil

### Requirement: ExError level mapping in Error method
The system SHALL use `errors.As` to detect `*exerror.ExError` in the `Error()` method. When detected, the mapping SHALL be: LevelFatal→LogLevelFatal, LevelUnexpected→LogLevelError, LevelExpected→LogLevelWarn, LevelSilent→no output (return immediately). For non-ExError errors, the level SHALL be LogLevelError.

#### Scenario: ExError with LevelFatal maps to LogLevelFatal
- **WHEN** `logger.Error("cat", exerror.New(..., exerror.LevelFatal, ...), nil)` is called
- **THEN** the log level sent to outputs is LogLevelFatal

#### Scenario: ExError with LevelExpected maps to LogLevelWarn
- **WHEN** `logger.Error("cat", exerror.New(..., exerror.LevelExpected, ...), nil)` is called
- **THEN** the log level sent to outputs is LogLevelWarn

#### Scenario: ExError with LevelSilent produces no output
- **WHEN** `logger.Error("cat", exerror.New(..., exerror.LevelSilent, ...), nil)` is called
- **THEN** no output's Log() is called

#### Scenario: Plain error maps to LogLevelError
- **WHEN** `logger.Error("cat", fmt.Errorf("oops"), nil)` is called
- **THEN** the log level sent to outputs is LogLevelError

### Requirement: LoggerOutput interface
The system SHALL define a `LoggerOutput` interface with two methods: `Log(data LoggerData)` and `Close() error`.

#### Scenario: Interface contract
- **WHEN** a type implements `Log(LoggerData)` and `Close() error`
- **THEN** it satisfies the LoggerOutput interface

### Requirement: Logger dispatches to all outputs
The system SHALL call `Log(data)` on every registered output for each write, regardless of log level. Filtering SHALL be the responsibility of each output.

#### Scenario: All outputs receive log data
- **WHEN** `logger.Info("cat", nil)` is called with two outputs registered
- **THEN** both outputs' `Log()` method is called with the same LoggerData
