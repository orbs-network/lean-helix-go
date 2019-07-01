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

func (l *Latch) ReturnWhenLatchIsResumed(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case l.pauseChannel <- true:
	}

	l.log.Debug("ReturnWhenLatchIsResumed() waiting for latch to resume")
	select {
	case <-ctx.Done():
		return
	case <-l.resumeChannel:
		l.log.Debug("ReturnWhenLatchIsResumed() latch has resumed")
	}
}

func (l *Latch) ReturnWhenLatchIsPaused(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-l.pauseChannel:
		l.log.Debug("ReturnWhenLatchIsPaused() latch has paused")
	}
}

func (l *Latch) Resume(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case l.resumeChannel <- true:
		l.log.Debug("Resume() resuming latch")
	}
}
