package logger

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type SilentLogger struct {
}

func NewSilentLogger() interfaces.Logger {
	return &SilentLogger{}
}

func (*SilentLogger) Debug(format string, args ...interface{}) {
}

func (*SilentLogger) Info(format string, args ...interface{}) {
}

func (*SilentLogger) Error(format string, args ...interface{}) {
}
