package logger

import (
	"fmt"
	"io"
	"time"
)

const (
	ansiReset     = "\033[0m"
	ansiDim       = "\033[2m"
	ansiCyan      = "\033[36m"
	ansiGreen     = "\033[32m"
	ansiYellow    = "\033[33m"
	ansiRed       = "\033[31m"
	ansiRedBgWhite = "\033[41;37m"
)

type ConsoleOutput struct {
	levels map[LogLevel]struct{}
	colors map[LogLevel]string
	writer io.Writer
}

func NewConsoleOutput(levels ...LogLevel) *ConsoleOutput {
	levelMap := make(map[LogLevel]struct{}, len(levels))
	for _, l := range levels {
		levelMap[l] = struct{}{}
	}

	colors := map[LogLevel]string{
		LogLevelDebug:   ansiDim,
		LogLevelInfo:    ansiCyan,
		LogLevelSuccess: ansiGreen,
		LogLevelWarn:    ansiYellow,
		LogLevelError:   ansiRed,
		LogLevelFatal:   ansiRedBgWhite,
	}

	return &ConsoleOutput{
		levels: levelMap,
		colors: colors,
		writer: defaultWriter{},
	}
}

func (c *ConsoleOutput) WithColors(colors map[LogLevel]string) *ConsoleOutput {
	for k, v := range colors {
		c.colors[k] = v
	}
	return c
}

func (c *ConsoleOutput) WithWriter(w io.Writer) *ConsoleOutput {
	c.writer = w
	return c
}

func (c *ConsoleOutput) Log(data LoggerData) {
	if _, ok := c.levels[data.Level]; !ok {
		return
	}

	timeStr := data.Time.Format(time.RFC3339)
	line := fmt.Sprintf("%s,%d,%s,%s,%s,%s", timeStr, data.Level, data.Identify, data.Category, data.Position, data.Content)

	color, ok := c.colors[data.Level]
	if !ok {
		color = ansiReset
	}

	fmt.Fprintln(c.writer, color+line+ansiReset)
}

func (c *ConsoleOutput) Close() error {
	return nil
}

type defaultWriter struct{}

func (defaultWriter) Write(p []byte) (int, error) {
	return fmt.Print(string(p))
}

var _ io.Writer = defaultWriter{}
var _ LoggerOutput = (*ConsoleOutput)(nil)
