package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

type LogLevel int

const (
	LogLevelDebug   LogLevel = iota + 1
	LogLevelInfo
	LogLevelSuccess
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelSuccess:
		return "SUCCESS"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type LoggerData struct {
	Time     time.Time
	Identify string
	Category string
	Level    LogLevel
	Error    error
	Content  string
	Position string
	Stack    []byte
	PID      int
}

type LoggerOutput interface {
	Log(data LoggerData)
	Close() error
}

type errorMessageData struct {
	Code    string `json:"code" yaml:"code"`
	Name    string `json:"name" yaml:"name"`
	Message string `json:"message" yaml:"message"`
	Stack   []string `json:"stack" yaml:"stack"`
	Args    any     `json:"args,omitempty" yaml:"args,omitempty"`
}

func ErrorMessage(err error) any {
	var exErr *errorx.Error
	if errors.As(err, &exErr) {
		return errorMessageData{
			Code:    exErr.Code,
			Name:    exErr.Name,
			Message: exErr.Message,
			Stack:   extractErrorStack(err),
			Args:    exErr.Args,
		}
	}
	return errorMessageData{
		Message: err.Error(),
		Stack:   extractErrorStack(err),
	}
}

func extractErrorStack(err error) []string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	raw := string(buf[:n])

	var frames []string
	for _, line := range splitStackLines(raw) {
		frames = append(frames, line)
	}
	return frames
}

func splitStackLines(raw string) []string {
	var lines []string
	remaining := raw
	for len(remaining) > 0 {
		idx := indexOfNewline(remaining)
		if idx < 0 {
			break
		}
		line := remaining[:idx]
		remaining = remaining[idx+1:]
		if len(line) > 0 && line[0] == '\t' {
			lines = append(lines, line[1:])
		}
	}
	return lines
}

func indexOfNewline(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return i
		}
	}
	return -1
}

type Logger struct {
	identify string
	outputs  []LoggerOutput
}

func NewLogger(identify string) *Logger {
	return &Logger{
		identify: identify,
		outputs:  nil,
	}
}

func (l *Logger) AddOutput(output LoggerOutput) *Logger {
	l.outputs = append(l.outputs, output)
	return l
}

func (l *Logger) Debug(category string, content any) {
	l.write(LogLevelDebug, category, nil, content)
}

func (l *Logger) Info(category string, content any) {
	l.write(LogLevelInfo, category, nil, content)
}

func (l *Logger) Success(category string, content any) {
	l.write(LogLevelSuccess, category, nil, content)
}

func (l *Logger) Warn(category string, content any) {
	l.write(LogLevelWarn, category, nil, content)
}

func (l *Logger) Error(category string, err error, content any) {
	var exErr *errorx.Error
	if errors.As(err, &exErr) {
		switch exErr.Level {
		case errorx.LevelFatal:
			l.write(LogLevelFatal, category, err, content)
		case errorx.LevelExpected:
			l.write(LogLevelWarn, category, err, content)
		case errorx.LevelSilent:
			return
		default:
			l.write(LogLevelError, category, err, content)
		}
	} else {
		l.write(LogLevelError, category, err, content)
	}
}

func (l *Logger) Fatal(category string, err error, content any) {
	l.write(LogLevelFatal, category, err, content)
}

func (l *Logger) Close() error {
	var firstErr error
	for _, output := range l.outputs {
		if err := output.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (l *Logger) write(level LogLevel, category string, err error, content any) {
	now := time.Now()
	pid := os.Getpid()

	contentStr := ""
	if content != nil {
		b, jsonErr := json.Marshal(content)
		if jsonErr == nil {
			contentStr = string(b)
		}
	}

	_, file, line, ok := runtime.Caller(2)
	position := "unknown:?"
	if ok {
		position = fmt.Sprintf("%s:%d", path.Base(file), line)
	}

	var stack []byte
	if level >= LogLevelError {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stack = buf[:n]
	}

	data := LoggerData{
		Time:     now,
		Identify: l.identify,
		Category: category,
		Level:    level,
		Error:    err,
		Content:  contentStr,
		Position: position,
		Stack:    stack,
		PID:      pid,
	}

	for _, output := range l.outputs {
		output.Log(data)
	}
}
