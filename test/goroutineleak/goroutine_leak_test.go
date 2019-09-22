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
	runtime.MemProfileRate = 1
	before, _ := os.Create("/tmp/lh-mem-before.prof")
	defer before.Close()
	after, _ := os.Create("/tmp/lh-mem-after4.prof")
	defer after.Close()

	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
	pprof.WriteHeapProfile(before)

	test.WithContext(func(ctx context.Context) {
		net1 := network.ATestNetworkBuilder(21).
			//LogToConsole(t). // This is a very long test, running with logs lets you view progress
			Build(ctx).
			StartConsensus(ctx)
		net1.WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx, 70, 1)

		//net2 := network.ATestNetworkBuilder(4).
		//	LogToConsole(t).
		//	Build(ctx).
		//	StartConsensus(ctx)
		//net2.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 20)
	})

	time.Sleep(20 * time.Millisecond) // give goroutines time to terminate

	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
	pprof.WriteHeapProfile(after)

}

// TODO Incorrect test, it should be updated to take into consideration there are now mainloop and workerloop goroutines
func TestGoroutinesLeaks(t *testing.T) {
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	heapBefore := int64(memBefore.HeapAlloc)
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

	runtime.ReadMemStats(&memAfter)
	heapAfter := int64(memAfter.HeapAlloc)
	numGoroutineAfter := runtime.NumGoroutine()
	pprof.Lookup("goroutine").WriteTo(after, 1)
	t.Logf("Memory: Before=%d After=%d", heapBefore, heapAfter)

	require.Equal(t, numGoroutineBefore, numGoroutineAfter, "number of goroutines should be equal, compare /tmp/leanhelix-goroutine-shutdown-before.out and /tmp/leanhelix-goroutine-shutdown-after.out to see stack traces of the leaks")
}
