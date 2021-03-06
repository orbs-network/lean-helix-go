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
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/scribe/log"
	"math"
	"time"
)

const PRINT_TIMESTAMP = true
const FORCE_STDOUT = false

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
	state          *state.State
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

const (
	LOG_LEVEL_DEBUG int = 1
	LOG_LEVEL_INFO  int = 2
	LOG_LEVEL_ERROR int = 3
)

func (l *lhLogger) log(level int, format string, args ...interface{}) {
	var f func(format string, args ...interface{})
	switch level {
	case LOG_LEVEL_INFO:
		f = l.externalLogger.Info
	case LOG_LEVEL_ERROR:
		f = l.externalLogger.Error
	default:
		f = l.externalLogger.Debug
	}

	lc := LC(l.state.Height(), l.state.View(), l.config.Membership.MyMemberId())
	s := fmt.Sprintf(format, args...)
	if FORCE_STDOUT {
		fmt.Printf(fmt.Sprintf("%s%s %s\n", nowISO(), lc, s))
	} else {
		f(fmt.Sprintf("%s%s %s", nowISO(), lc, s))
	}
}

func (l *lhLogger) Debug(format string, args ...interface{}) {
	l.log(LOG_LEVEL_DEBUG, format, args...)
}

func (l *lhLogger) Info(format string, args ...interface{}) {
	l.log(LOG_LEVEL_INFO, format, args...)
}

func (l *lhLogger) Error(format string, args ...interface{}) {
	l.log(LOG_LEVEL_ERROR, format, args...)
}

func (l *lhLogger) ConsensusTrace(msg string, err error, fields ...*log.Field) {
	fields = append(fields, log.Uint64("block-height", uint64(l.state.Height())), log.Uint64("view", uint64(l.state.View())))
	if err != nil {
		fields = append(fields, log.Error(err))
	}
	l.externalLogger.ConsensusTrace(msg, fields...)
}

func NewLhLogger(config *interfaces.Config, state *state.State) LHLogger {
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
	ConsensusTrace(msg string, err error, fields ...*log.Field)
}

func LC(h primitives.BlockHeight, v primitives.View, id primitives.MemberId) *_LC {
	return &_LC{
		h:  h,
		v:  v,
		id: id,
	}
}

type MessageLog struct {
	MessageType protocol.MessageType
	BlockHash   primitives.BlockHash
}

func blockHashToStr(blockHash primitives.BlockHash) string {
	if blockHash == nil {
		return "none"
	}
	blockHashStr := hex.EncodeToString(blockHash)
	if len(blockHashStr) < 6 {
		return blockHashStr
	}
	return blockHashStr[:6]
}

type MemberMessagesLog struct {
	MemberId primitives.MemberId
	Messages []MessageLog
}

func (mml *MemberMessagesLog) String() string {
	memberMessages := fmt.Sprintf("[ memberId=%s,  member-messages::", MemberIdToStr(mml.MemberId))
	for _, message := range mml.Messages {
		memberMessages += fmt.Sprintf(" (message-type=%s; block-hash=%s)", message.MessageType, blockHashToStr(message.BlockHash))
	}
	memberMessages += "] "

	return memberMessages
}

type MemberIdStr string

func ConvertMessagesToMemberMessagesLogs(messages []interface{}) string {

	memberMessages := make(map[MemberIdStr][]MessageLog)

	for _, message := range messages {
		var blockHash primitives.BlockHash
		var memberId primitives.MemberId
		var messageType protocol.MessageType

		switch message := message.(type) {
		case *interfaces.PreprepareMessage:
			messageType = message.MessageType()
			memberId = message.Content().Sender().MemberId()
			blockHash = message.Content().SignedHeader().BlockHash()

		case *interfaces.PrepareMessage:
			messageType = message.MessageType()
			memberId = message.Content().Sender().MemberId()
			blockHash = message.Content().SignedHeader().BlockHash()

		case *interfaces.CommitMessage:
			messageType = message.MessageType()
			memberId = message.Content().Sender().MemberId()
			blockHash = message.Content().SignedHeader().BlockHash()

		case *interfaces.ViewChangeMessage:
			messageType = message.MessageType()
			memberId = message.Content().Sender().MemberId()
			blockHash = nil
			if message.Block() != nil {
				blockHash = message.Content().SignedHeader().PreparedProof().PrepareBlockRef().BlockHash()
			}

		default:
			continue
		}
		memberIdStr := MemberIdStr(memberId)
		memberMessages[memberIdStr] = append(memberMessages[memberIdStr],
			MessageLog{
				MessageType: messageType,
				BlockHash:   blockHash,
			})
	}

	if len(memberMessages) > 0 {
		output := ""
		for memberIdStr, messages := range memberMessages {
			memberMessagesStr := fmt.Sprintf("[ memberId=%s,  member-messages::", MemberIdToStr(primitives.MemberId(memberIdStr)))
			for _, message := range messages {
				memberMessagesStr += fmt.Sprintf(" (message-type=%s; block-hash=%s)", message.MessageType, blockHashToStr(message.BlockHash))
			}
			memberMessagesStr += "] "
			output += memberMessagesStr
		}
		return output
	} else {
		return "No messages"
	}
}
