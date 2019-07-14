// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Latch struct {
	log           interfaces.Logger
	pauseChannel  chan bool
	resumeChannel chan bool
}

func NewLatch() *Latch {
	return &Latch{
		log:           logger.NewConsoleLogger(),
		pauseChannel:  make(chan bool),
		resumeChannel: make(chan bool),
	}
}

func (l *Latch) ReturnWhenLatchIsResumed(ctx context.Context, memberId primitives.MemberId) {
	select {
	case <-ctx.Done():
		return
	case l.pauseChannel <- true:
	}

	l.log.Debug("ID=%s Latch.ReturnWhenLatchIsResumed() waiting for latch to resume", memberId)
	select {
	case <-ctx.Done():
		return
	case <-l.resumeChannel:
		l.log.Debug("ID=%s Latch.ReturnWhenLatchIsResumed() latch has resumed", memberId)
	}
}

func (l *Latch) ReturnWhenLatchIsPaused(ctx context.Context, memberId primitives.MemberId) {
	select {
	case <-ctx.Done():
		return
	case <-l.pauseChannel:
		l.log.Debug("ID=%s Latch.ReturnWhenLatchIsPaused() latch has paused", memberId)
	}
}

func (l *Latch) Resume(ctx context.Context, memberId primitives.MemberId) {
	select {
	case <-ctx.Done():
		return
	case l.resumeChannel <- true:
		l.log.Debug("ID=%s Latch.Resume() resuming latch", memberId)
	}
}
