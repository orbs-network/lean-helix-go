// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package logger

import (
	"encoding/hex"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"math"
	"time"
)

const PRINT_TIMESTAMP = true

type _LC struct {
	h  primitives.BlockHeight
	v  primitives.View
	id primitives.MemberId
}

func (lc *_LC) String() string {
	if lc == nil {
		return ""
	}

	var hStr string
	if lc.h == math.MaxUint64 {
		hStr = "X"
	} else {
		hStr = fmt.Sprintf("%d", lc.h)
	}

	var vStr string
	if lc.v == math.MaxUint64 {
		vStr = "X"
	} else {
		vStr = fmt.Sprintf("%d", lc.v)
	}

	return fmt.Sprintf("H=%s V=%s ID=%s", hStr, vStr, MemberIdToStr(lc.id))
}

type lhLogger struct {
	config         *interfaces.Config
	state          state.State
	externalLogger interfaces.Logger
}

func (l *lhLogger) ExternalLogger() interfaces.Logger {
	return l.externalLogger
}

func nowISO() string {
	// Full ISO8601 is "2006-01-02T15:04:05.000Z"
	if PRINT_TIMESTAMP {
		return time.Now().Format("15:04:05.000Z ")
	} else {
		return ""
	}
}

func (l *lhLogger) Debug(format string, args ...interface{}) {
	//fmt.Printf(fmt.Sprintf("%s%s %s\n", nowISO(), LC(l.state.Height(), l.state.View(), l.config.Membership.MyMemberId()), format), args...)
	l.externalLogger.Debug(fmt.Sprintf("%s%s %s", nowISO(), LC(l.state.Height(), l.state.View(), l.config.Membership.MyMemberId()), format), args...)
}

func (l *lhLogger) Info(format string, args ...interface{}) {
	//fmt.Printf(fmt.Sprintf("%s%s %s\n", nowISO(), LC(l.state.Height(), l.state.View(), l.config.Membership.MyMemberId()), format), args...)
	l.externalLogger.Info(fmt.Sprintf("%s%s %s", nowISO(), LC(l.state.Height(), l.state.View(), l.config.Membership.MyMemberId()), format), args...)
}

func (l *lhLogger) Error(format string, args ...interface{}) {
	//fmt.Printf(fmt.Sprintf("%s%s %s\n", nowISO(), LC(l.state.Height(), l.state.View(), l.config.Membership.MyMemberId()), format), args...)
	l.externalLogger.Error(fmt.Sprintf("%s%s %s", nowISO(), LC(l.state.Height(), l.state.View(), l.config.Membership.MyMemberId()), format), args...)
}

func NewLhLogger(config *interfaces.Config, state state.State) LHLogger {
	var logger interfaces.Logger
	if config.Logger == nil {
		logger = NewSilentLogger()
	} else {
		logger = config.Logger
	}
	return &lhLogger{
		config:         config,
		state:          state,
		externalLogger: logger,
	}
}

func MemberIdToStr(memberId primitives.MemberId) string {
	if memberId == nil {
		return ""
	}
	memberIdStr := hex.EncodeToString(memberId)
	if len(memberIdStr) < 6 {
		return memberIdStr
	}
	return memberIdStr[:6]
}

type LHLogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	ExternalLogger() interfaces.Logger
}

func LC(h primitives.BlockHeight, v primitives.View, id primitives.MemberId) *_LC {
	return &_LC{
		h:  h,
		v:  v,
		id: id,
	}
}
