## ADDED Requirements

### Requirement: ConsoleOutput construction
The system SHALL provide `NewConsoleOutput(levels ...LogLevel) *ConsoleOutput`. It SHALL store the provided levels as a `map[LogLevel]struct{}` whitelist for filtering.

#### Scenario: ConsoleOutput created with specific levels
- **WHEN** `NewConsoleOutput(LogLevelInfo, LogLevelError)` is called
- **THEN** the ConsoleOutput accepts only Info and Error levels

#### Scenario: ConsoleOutput with no levels accepts nothing
- **WHEN** `NewConsoleOutput()` is called with no arguments
- **THEN** the ConsoleOutput filters out all log levels

### Requirement: ConsoleOutput level filtering
The ConsoleOutput SHALL only output log entries whose Level exists in its whitelist. Entries with a Level not in the whitelist SHALL be silently discarded.

#### Scenario: Level in whitelist is output
- **WHEN** `ConsoleOutput.Log(data)` is called with data.Level = LogLevelInfo and LogLevelInfo is in the whitelist
- **THEN** the entry is formatted and printed to stdout

#### Scenario: Level not in whitelist is discarded
- **WHEN** `ConsoleOutput.Log(data)` is called with data.Level = LogLevelDebug and LogLevelDebug is NOT in the whitelist
- **THEN** nothing is printed and the method returns immediately

### Requirement: ConsoleOutput default colors
The system SHALL provide default ANSI color codes per LogLevel: Debug→dim(\033[2m), Info→cyan(\033[36m), Success→green(\033[32m), Warn→yellow(\033[33m), Error→red(\033[31m), Fatal→red background white foreground(\033[41;37m).

#### Scenario: Default color for Info level
- **WHEN** a ConsoleOutput is created without custom colors and logs an Info level entry
- **THEN** the output is wrapped with \033[36m (cyan) and terminated with \033[0m

#### Scenario: Default color for Fatal level
- **WHEN** a ConsoleOutput is created without custom colors and logs a Fatal level entry
- **THEN** the output is wrapped with \033[41;37m (red bg, white fg) and terminated with \033[0m

### Requirement: ConsoleOutput custom colors
The system SHALL allow overriding colors per LogLevel via `WithColors(colors map[LogLevel]string)` option or constructor parameter. Only the provided levels in the map SHALL be overridden; unspecified levels SHALL use defaults.

#### Scenario: Custom color overrides default
- **WHEN** a ConsoleOutput is created with custom color map `{LogLevelInfo: "\033[35m"}`
- **THEN** Info level entries use \033[35m instead of cyan, while other levels retain defaults

### Requirement: ConsoleOutput format
The system SHALL output each log entry in the format: `timeString,LEVEL,identify,category,position,content`, wrapped in the appropriate ANSI color code and terminated with a reset code. The timeString SHALL be formatted as `YYYY-MM-DD HH:mm:ss.SSS`.

#### Scenario: Formatted output matches pattern
- **WHEN** a ConsoleOutput logs data with Time="2026-05-05 14:30:22.123", Level=INFO, Identify="svc", Category="rpc", Position="conn.go:42", Content=`{"ok":true}`
- **THEN** the printed line is `\033[36m2026-05-05 14:30:22.123,INFO,svc,rpc,conn.go:42,{"ok":true}\033[0m`

### Requirement: ConsoleOutput Close
The ConsoleOutput `Close()` method SHALL be a no-op and return nil.

#### Scenario: Close returns nil
- **WHEN** `ConsoleOutput.Close()` is called
- **THEN** nil is returned with no side effects
