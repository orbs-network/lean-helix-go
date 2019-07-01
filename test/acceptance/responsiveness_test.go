package acceptance

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNodeSyncNotProcessedWhileRunningLongOperation(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		node0 := net.Nodes[0]
		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		fmt.Println("Calling WaitForNodeToRequestNewBlock")

		net.WaitForNodeToRequestNewBlock(ctx, net.Nodes[0])

		// Send NodeSync
		node0.Sync(ctx, nil, nil, nil)

		// Verify OnNewConsensusRound is called (this should fail)

		fmt.Println("Called WaitForNodeToRequestNewBlock, waiting 1s")

		time.Sleep(1 * time.Second)
		fmt.Println("Waited for 1s, calling ResumeNodeRequestNewBlock")
		net.ResumeNodeRequestNewBlock(ctx, net.Nodes[0])
		fmt.Println("Called ResumeNodeRequestNewBlock")

		require.True(t, net.WaitForAllNodesToCommitTheSameBlock(ctx))
	})

	// Start consensus
	// Run a paused CreateBlock

}

func TestElectionNotProcessedWhileRunningLongOperation(t *testing.T) {
	// Start consensus
	// Run a paused CreateBlock
	// Send ElectionTrigger
	// Verify SetView is called (this should fail)
}

func TestWorkerContextCancellation(t *testing.T) {

}
