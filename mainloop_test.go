package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"testing"
	"time"
)

func TestRunEndsAfterGoroutinesEnd(t *testing.T) {

	shutdownContext, cancelTimer := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelTimer()

	ctx, cancelGoRoutines := context.WithCancel(context.Background())
	mainLoop := NewLeanHelix(mocks.NewMockConfigSimple(), nil, nil)
	mainLoop.Run(ctx)
	time.Sleep(100 * time.Millisecond) // TODO replace with latch that fires after both goroutines have started?
	cancelGoRoutines()
	mainLoop.WaitUntilShutdown(shutdownContext)

	select {
	case <-ctx.Done():
	// ok
	case <-shutdownContext.Done():
		t.Fatalf("system did not shut down in a timely manner")
	}
}
