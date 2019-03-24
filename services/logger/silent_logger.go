// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

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
