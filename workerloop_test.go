package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"sync"
	"testing"
	"time"
)

func DummyWorkerConfig() *interfaces.Config {
	return &interfaces.Config{
		InstanceId:      123,
		Communication:   nil,
		Membership:      mocks.NewMockMembership(primitives.MemberId{0, 1, 2}, nil, true),
		BlockUtils:      nil,
		KeyManager:      nil,
		ElectionTrigger: nil,
		Storage:         nil,
		Logger:          logger.NewSilentLogger(),
		MsgChanBufLen:   10,
	}
}

func TestWorkerLoopReturnsOnMainContextCancellation(t *testing.T) {

	test.WithContext(func(ctx context.Context) {

		mainCtx, mainCancel := context.WithCancel(ctx)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		cfg := DummyWorkerConfig()
		workerLoop := NewWorkerLoop(cfg, LoggerToLHLogger(cfg.Logger), nil)
		go func() {
			workerLoop.Run(mainCtx)
			wg.Done()
		}()
		mainCancel()

		test.FailIfNotDoneByTimeout(t, wg, 1*time.Second, "Main context was cancelled but worker loop did not return by timeout")
	})
}

// Write test that teases out Worker structure of loop with worker context that is cancelled on NodeSync and Election

func TestWorkerContextPropagatedToCancellableOperationsInWorkerLoop(t *testing.T) {

}
