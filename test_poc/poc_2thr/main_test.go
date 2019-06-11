package poc_2thr

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// See README.md for POC docs

func TestMainFlow(t *testing.T) {

	d := &Config{
		CancelTestAfter:      2000 * time.Millisecond,
		WaitAfterCancelTest:  500 * time.Millisecond,
		CreateBlock:          500 * time.Millisecond,
		ValidateBlock:        500 * time.Millisecond,
		MessageChannelBufLen: 10,
	}

	Run(d)
}

func TestCancelContextReturnsImmediately(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	go myfunc(ctx)
	time.Sleep(500 * time.Millisecond)
	startTime := time.Now()
	fmt.Printf("%s Cancelling\n", time.Now())
	cancel()
	fmt.Printf("%s Cancelled\n", time.Now())
	endTime := time.Now()
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("%s TEST END\n", time.Now())
	// If cancel() were not to return immediately, this test would fail
	// Note that cancelling does not finish the goroutine immediately, but only when the goroutine checks ctx.Done()
	require.True(t, endTime.Sub(startTime) < 50*time.Millisecond)

}

func myfunc(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("%s Done\n", time.Now())
			return

		default:
			fmt.Printf("%s Sleeping start\n", time.Now())
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("%s Sleeping end\n", time.Now())
		}
	}
}
