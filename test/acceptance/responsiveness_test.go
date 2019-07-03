package acceptance

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNodeSynHangsWhileRunningLongOperation(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		net := network.ATestNetwork(4, t.Logf, block1, block2)
		node0 := net.Nodes[0]

		net.SetNodesToPauseOnRequestNewBlock()

		node0.StartConsensus(ctx)
		net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)

		fmt.Println("Node is paused on create block..")

		latestBlock := node0.GetLatestBlock()
		newBlock := mocks.ABlock(latestBlock)
		require.Equal(t, primitives.BlockHeight(1), node0.GetCurrentHeight())
		net.SetNodesToPauseOnHandleUpdateState()

		time.AfterFunc(1*time.Second, func() {
			t.Errorf("Waited too long for Sync to complete")
			t.FailNow()
		})

		node0.SyncWithoutProof(ctx, newBlock, nil)
		net.ReturnWhenNodesPauseOnUpdateState(ctx, node0)
	})
}
