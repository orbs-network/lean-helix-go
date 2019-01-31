package logger

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type ConsoleLogger struct {
	level LogLevel
}

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

func NewConsoleLogger() interfaces.Logger {
	return &ConsoleLogger{
		level: LEVEL_DEBUG,
	}
}
