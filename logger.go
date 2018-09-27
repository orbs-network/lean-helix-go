package leanhelix

import "fmt"

type LogLevel int

const LEVEL_DEBUG LogLevel = 0
const LEVEL_INFO LogLevel = 1
const LEVEL_ERROR LogLevel = 2

type Logger interface {
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Error(format string, a ...interface{})
}

type ConsoleLogger struct {
	id    string
	level LogLevel
}

func (l *ConsoleLogger) Debug(format string, a ...interface{}) {
	if l.level > LEVEL_DEBUG {
		return
	}
	stdout("%s "+format, "LEVEL_DEBUG", a)
}

func (l *ConsoleLogger) Info(format string, a ...interface{}) {
	if l.level > LEVEL_INFO {
		return
	}
	stdout("%s "+format, "LEVEL_INFO", a)
}

func (l *ConsoleLogger) Error(format string, a ...interface{}) {
	if l.level > LEVEL_ERROR {
		return
	}
	stdout("%s "+format, "LEVEL_ERROR", a)
}

func stdout(format string, a ...interface{}) {
	fmt.Printf(format, a)
}

type SilentLogger struct {
}

func NewSilentLogger() *SilentLogger {
	return &SilentLogger{}
}

func NewConsoleLogger(id string) *ConsoleLogger {
	return &ConsoleLogger{
		id:    id,
		level: LEVEL_DEBUG,
	}
}

func (*SilentLogger) Debug(format string, a ...interface{}) {
}

func (*SilentLogger) Info(format string, a ...interface{}) {
}

func (*SilentLogger) Error(format string, a ...interface{}) {
}
