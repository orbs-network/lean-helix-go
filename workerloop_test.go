package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/electiontrigger"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"sync"
	"testing"
	"time"
)

// TODO This file is a candidate for removal - impractical to unit-test it

func DummyWorkerConfig() *interfaces.Config {
	return &interfaces.Config{
		InstanceId:            123,
		Communication:         nil,
		Membership:            mocks.NewFakeMembership(primitives.MemberId{0, 1, 2}, nil, nil, true),
		BlockUtils:            nil,
		KeyManager:            nil,
		ElectionTimeoutOnV0:   10 * time.Millisecond,
		OnElectionCB:          nil,
		Storage:               nil,
		Logger:                logger.NewSilentLogger(),
		MsgChanBufLen:         10,
		UpdateStateChanBufLen: 10,
		ElectionChanBufLen:    0,
	}
}

func TestWorkerLoopReturnsOnMainContextCancellation(t *testing.T) {

	test.WithContext(func(ctx context.Context) {

		mainCtx, mainCancel := context.WithCancel(ctx)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		cfg := DummyWorkerConfig()
		s := state.NewState()
		electionTrigger := Electiontrigger.NewTimerBasedElectionTrigger(cfg.ElectionTimeoutOnV0, nil)
		workerLoop := NewWorkerLoop(s, cfg, logger.NewLhLogger(cfg, s), electionTrigger, nil, nil)
		go func() {
			workerLoop.Run(mainCtx)
			wg.Done()
		}()
		mainCancel()

		test.FailIfNotDoneByTimeout(t, wg, 1*time.Second, "Main context was canceled but worker loop did not return by timeout")
	})
}

// Write test that teases out Worker structure of loop with worker context that is canceled on NodeSync and Election

func TestWorkerContextPropagatedToCancellableOperationsInWorkerLoop(t *testing.T) {

}
