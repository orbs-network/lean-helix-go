package leanhelix

import "fmt"

type LogLevel int

const LEVEL_DEBUG LogLevel = 0
const LEVEL_INFO LogLevel = 1
const LEVEL_ERROR LogLevel = 2

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type ConsoleLogger struct {
	id    string
	level LogLevel
}

type LogData map[string]string

// TODO Impl Stringer for LogData by printing each property, or use JSON.unmarshal or something

func (l *ConsoleLogger) Debug(format string, args ...interface{}) {
	if l.level > LEVEL_DEBUG {
		return
	}
	stdout(fmt.Sprintf("[DEBUG] - %s\n", format), args...)
}

func (l *ConsoleLogger) Info(format string, args ...interface{}) {
	if l.level > LEVEL_INFO {
		return
	}
	stdout(fmt.Sprintf("[INFO ] - %s\n", format), args...)
}

func (l *ConsoleLogger) Error(format string, args ...interface{}) {
	if l.level > LEVEL_ERROR {
		return
	}
	stdout(fmt.Sprintf("*ERROR* - %s\n", format), args...)
}

func stdout(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

type SilentLogger struct {
}

func NewSilentLogger() Logger {
	return &SilentLogger{}
}

func NewConsoleLogger(id string) Logger {
	return &ConsoleLogger{
		id:    id,
		level: LEVEL_DEBUG,
	}
}

func (*SilentLogger) Debug(format string, args ...interface{}) {
}

func (*SilentLogger) Info(format string, args ...interface{}) {
}

func (*SilentLogger) Error(format string, args ...interface{}) {
}
