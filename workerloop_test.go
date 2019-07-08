package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test"
	"sync"
	"testing"
	"time"
)

func TestWorkerLoopReturnsOnMainContextCancellation(t *testing.T) {

	test.WithContext(func(ctx context.Context) {

		mainCtx, cancel := context.WithCancel(ctx)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		workerLoop := NewWorkerLoop()
		go func() {
			workerLoop.Start(mainCtx)
			wg.Done()
		}()
		cancel()

		test.FailIfNotDoneByTimeout(t, wg, 1*time.Second, "Main context was cancelled but worker loop did not return by timeout")
	})
}
