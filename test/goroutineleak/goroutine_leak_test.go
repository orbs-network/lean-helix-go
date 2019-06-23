// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

//+build goroutineleak

package goroutineleak

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
)

func test2HeavyNetworks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net1 := network.ATestNetwork(21).StartConsensus(ctx)
		for i := 0; i < 100; i++ {
			net1.WaitForAllNodesToCommitTheSameBlock(ctx)
		}

		net2 := network.ATestNetwork(31).StartConsensus(ctx)
		for i := 0; i < 100; i++ {
			net2.WaitForAllNodesToCommitTheSameBlock(ctx)
		}
	})
}

func TestGoroutinesLeaks(t *testing.T) {
	before, _ := os.Create("/tmp/leanhelix-goroutine-shutdown-before.out")
	defer before.Close()
	after, _ := os.Create("/tmp/leanhelix-goroutine-shutdown-after.out")
	defer after.Close()

	numGoroutineBefore := runtime.NumGoroutine()
	pprof.Lookup("goroutine").WriteTo(before, 1)

	t.Run("test2HeavyNetworks", test2HeavyNetworks)

	time.Sleep(200 * time.Millisecond) // give goroutines time to terminate
	runtime.GC()
	time.Sleep(200 * time.Millisecond) // give goroutines time to terminate

	numGoroutineAfter := runtime.NumGoroutine()
	pprof.Lookup("goroutine").WriteTo(after, 1)

	require.Equal(t, numGoroutineBefore, numGoroutineAfter, "number of goroutines should be equal, compare /tmp/leanhelix-goroutine-shutdown-before.out and /tmp/leanhelix-goroutine-shutdown-after.out to see stack traces of the leaks")
}
