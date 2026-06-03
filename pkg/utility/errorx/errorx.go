package errorx

import (
	"fmt"
)

type ErrorLevel int

const (
	LevelFatal      ErrorLevel = -1
	LevelUnexpected ErrorLevel = 1
	LevelExpected   ErrorLevel = 2
	LevelSilent     ErrorLevel = 3
)

func (l ErrorLevel) String() string {
	switch l {
	case LevelFatal:
		return "FATAL"
	case LevelUnexpected:
		return "UNEXPECTED"
	case LevelExpected:
		return "EXPECTED"
	case LevelSilent:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

type Error struct {
	Code    string     `json:"code"`
	Level   ErrorLevel `json:"level"`
	Name    string     `json:"name"`
	Message string     `json:"message"`
	Args    any        `json:"args,omitempty"`
	Err     error      `json:"err,omitempty"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s (cause: %v)", e.Level.String(), e.Name, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Level.String(), e.Name, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code string, level ErrorLevel, name string, msg string, args any) *Error {
	return &Error{
		Code:    code,
		Level:   level,
		Name:    name,
		Message: msg,
		Args:    args,
	}
}

func Wrap(err error, code string, level ErrorLevel, name string, msg string, args any) *Error {
	return &Error{
		Code:    code,
		Level:   level,
		Name:    name,
		Message: msg,
		Args:    args,
		Err:     err,
	}
}
