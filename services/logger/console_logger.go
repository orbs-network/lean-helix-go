// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package logger

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/scribe/log"
)

// Deprecated; use LHLogger wrapping Scribe
type ConsoleLogger struct {
	level LogLevel
	uid   string
}

func (l *ConsoleLogger) ConsensusTrace(format string, fields ...*log.Field) {
	l.Debug(format) // this is shit, but the whole class needs to go
}

func (l *ConsoleLogger) Debug(format string, args ...interface{}) {
	if l.level > LEVEL_DEBUG {
		return
	}
	stdout(fmt.Sprintf("[D|%s] - %s\n", l.uid, format), args...)
}

func (l *ConsoleLogger) Info(format string, args ...interface{}) {
	if l.level > LEVEL_INFO {
		return
	}
	stdout(fmt.Sprintf("[I|%s] - %s\n", l.uid, format), args...)
}

func (l *ConsoleLogger) Error(format string, args ...interface{}) {
	if l.level > LEVEL_ERROR {
		return
	}
	stdout(fmt.Sprintf("*E|%s* - %s\n", l.uid, format), args...)
}

func stdout(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func NewConsoleLogger(uid string) interfaces.Logger {
	return &ConsoleLogger{
		level: LEVEL_DEBUG,
		uid:   uid,
	}
}
