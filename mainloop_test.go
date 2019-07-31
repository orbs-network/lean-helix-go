package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"testing"
	"time"
)

// TODO Fix this test
func TestRunEndsAfterGoroutinesEnd(t *testing.T) {

	ctx, cancelGoRoutines := context.WithCancel(context.Background())

	mainloop := NewLeanHelix(mocks.NewMockConfig(), nil, nil)

	cancelWrapper := func() {
		cancelGoRoutines()
		t.Log("Canceled now")
	}
	timer := time.AfterFunc(200*time.Millisecond, cancelWrapper)
	timerBeforeCancelContext, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	timerAfterCancelContext, _ := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer func() {
		cancelGoRoutines()
		timer.Stop()
		t.Log("Stopped")
	}()
	runDone := make(chan struct{})
	go func() {
		t.Log("Start Run")
		startTime := time.Now()
		mainloop.Run(ctx)
		t.Logf("End Run: %fs", time.Now().Sub(startTime).Seconds())
		close(runDone)
	}()

	select {
	case <-timerBeforeCancelContext.Done():
		t.Log("timerBeforeCancelContext.Done")
	case <-runDone:
		t.Fatal("Run ended before its goroutines ended")
	}

	select {
	case <-timerAfterCancelContext.Done():
		t.Fatal("Context canceled but Run() did not immediately end")
	case <-runDone:
		t.Log("runDone")
		return
	}
	t.Fatal("Shouldn't reach here")
}
