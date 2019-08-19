// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Latch struct {
	log           interfaces.Logger
	pauseChannel  chan bool
	resumeChannel chan bool
}

func NewLatch(logger interfaces.Logger) *Latch {
	return &Latch{
		log:           logger,
		pauseChannel:  make(chan bool),
		resumeChannel: make(chan bool),
		//paused:        false,
	}
}

func (l *Latch) WaitOnPauseThenWaitOnResume(ctx context.Context, memberId primitives.MemberId) {

	l.log.Debug("ID=%s Latch.WaitOnPauseThenWaitOnResume() start, blocked on writing to Pause channel", memberId)
	select {
	case <-ctx.Done():
		l.log.Debug("ID=%s Latch.WaitOnPauseThenWaitOnResume() ctx.Done (before Pause)", memberId)
		return
	case l.pauseChannel <- true:
		l.log.Debug("ID=%s Latch.WaitOnPauseThenWaitOnResume() wrote to paused latch", memberId)
	}

	l.log.Debug("ID=%s Latch.WaitOnPauseThenWaitOnResume() blocked on writing to resume channel", memberId)
	select {
	case <-ctx.Done():
		l.log.Debug("ID=%s Latch.WaitOnPauseThenWaitOnResume() ctx.Done (before Resume)", memberId)
		return
	case <-l.resumeChannel:
		l.log.Debug("ID=%s Latch.WaitOnPauseThenWaitOnResume() read from resume channel", memberId)
	}
}

func (l *Latch) ReturnWhenLatchIsPaused(ctx context.Context, memberId primitives.MemberId) {
	l.log.Debug("ID=%s Latch.ReturnWhenLatchIsPaused() start, blocked on reading from pause channel", memberId)
	select {
	case <-ctx.Done():
		l.log.Debug("ID=%s Latch.ReturnWhenLatchIsPaused() ctx.Done", memberId)
		return
	case <-l.pauseChannel:
		l.log.Debug("ID=%s Latch.ReturnWhenLatchIsPaused() read from Pause channel", memberId)
	}
}

func (l *Latch) Resume(ctx context.Context, memberId primitives.MemberId) {
	l.log.Debug("ID=%s Latch.Resume() start, blocked on reading from Resume channel", memberId)
	select {
	case <-ctx.Done():
		l.log.Debug("ID=%s Latch.Resume() ctx.Done", memberId)
		return
	case l.resumeChannel <- true:
		l.log.Debug("ID=%s Latch.Resume() wrote to Resume channel", memberId)
	}
}
