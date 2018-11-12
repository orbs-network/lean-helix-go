package test

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	WithContext(func(ctx context.Context) {
		et := lh.NewTimerBasedElectionTrigger(10 * time.Millisecond)
		resultContext := et.CreateElectionContext(ctx, 0)
		time.Sleep(time.Duration(15) * time.Millisecond)
		require.Error(t, resultContext.Err())
	})
}

func TestIgnoreSameView(t *testing.T) {
	WithContext(func(ctx context.Context) {
		et := lh.NewTimerBasedElectionTrigger(30 * time.Millisecond)

		resultContext := et.CreateElectionContext(ctx, 0)
		time.Sleep(time.Duration(10) * time.Millisecond)
		resultContext = et.CreateElectionContext(ctx, 0)
		time.Sleep(time.Duration(10) * time.Millisecond)
		resultContext = et.CreateElectionContext(ctx, 0)

		require.NoError(t, resultContext.Err())

		time.Sleep(time.Duration(20) * time.Millisecond)
		resultContext = et.CreateElectionContext(ctx, 0)

		require.Error(t, resultContext.Err())
	})
}

func TestViewChange(t *testing.T) {
	WithContext(func(ctx context.Context) {
		et := lh.NewTimerBasedElectionTrigger(20 * time.Millisecond)

		resultContext := et.CreateElectionContext(ctx, 0) // 2 ** 0 * 20 = 20
		time.Sleep(time.Duration(10) * time.Millisecond)

		resultContext = et.CreateElectionContext(ctx, 1) // 2 ** 1 * 20 = 40
		time.Sleep(time.Duration(30) * time.Millisecond)

		resultContext = et.CreateElectionContext(ctx, 2) // 2 ** 2 * 20 = 80
		time.Sleep(time.Duration(70) * time.Millisecond)

		resultContext = et.CreateElectionContext(ctx, 3) // 2 ** 3 * 20 = 160

		require.NoError(t, resultContext.Err())
	})
}

func TestViewPowTimeout(t *testing.T) {
	WithContext(func(ctx context.Context) {
		et := lh.NewTimerBasedElectionTrigger(10 * time.Millisecond)

		resultContext := et.CreateElectionContext(ctx, 2) // 2 ** 2 * 20 = 40
		time.Sleep(time.Duration(30) * time.Millisecond)
		require.NoError(t, resultContext.Err())
		time.Sleep(time.Duration(30) * time.Millisecond)
		require.Error(t, resultContext.Err())
	})
}
