package leanhelix

import "fmt"

type LogLevel int

const LEVEL_DEBUG LogLevel = 0
const LEVEL_INFO LogLevel = 1
const LEVEL_ERROR LogLevel = 2

type Logger interface {
	Debug(msg string, data *LogData)
	Info(msg string, data *LogData)
	Error(msg string, data *LogData)
}

type ConsoleLogger struct {
	id    string
	level LogLevel
}

type LogData map[string]string

// TODO Impl Stringer for LogData by printing each property, or use JSON.unmarshal or something

func (l *ConsoleLogger) Debug(msg string, data *LogData) {
	if l.level > LEVEL_DEBUG {
		return
	}
	stdout("[DEBUG] - %s %s", msg, data)
}

func (l *ConsoleLogger) Info(msg string, data *LogData) {
	if l.level > LEVEL_INFO {
		return
	}
	stdout("[INFO ] - %s %s", msg, data)
}

func (l *ConsoleLogger) Error(msg string, data *LogData) {
	if l.level > LEVEL_ERROR {
		return
	}
	stdout("*ERROR* - %s %s", msg, data)
}

func stdout(format string, args ...interface{}) {
	fmt.Printf(format, args)
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
