package logger

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
)

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
	externalLogger interfaces.Logger
}

func (l *lhLogger) ExternalLogger() interfaces.Logger {
	return l.externalLogger
}

func (l *lhLogger) Debug(lc *_LC, format string, args ...interface{}) {
	l.externalLogger.Debug(fmt.Sprintf("%s %s", lc, format), args...)
}

func (l *lhLogger) Info(lc *_LC, format string, args ...interface{}) {
	l.externalLogger.Info(fmt.Sprintf("%s %s", lc, format), args...)
}

func (l *lhLogger) Error(lc *_LC, format string, args ...interface{}) {
	l.externalLogger.Error(fmt.Sprintf("%s %s", lc, format), args...)
}

func NewLhLogger(externalLogger interfaces.Logger) LHLogger {
	return &lhLogger{
		externalLogger: externalLogger,
	}
}

func MemberIdToStr(memberId primitives.MemberId) string {
	if memberId == nil {
		return ""
	}
	return memberId.String()[:6]
}

type LHLogger interface {
	Debug(lc *_LC, format string, args ...interface{})
	Info(lc *_LC, format string, args ...interface{})
	Error(lc *_LC, format string, args ...interface{})
	ExternalLogger() interfaces.Logger
}

func LC(h primitives.BlockHeight, v primitives.View, id primitives.MemberId) *_LC {
	return &_LC{
		h:  h,
		v:  v,
		id: id,
	}
}
