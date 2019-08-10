// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/leaderelection"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodeSync_AllNodesReachSameHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.ATestNetworkBuilder(4, block1, block2, block3).
			//LogToConsole(t).
			Build(ctx)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		// closing node3's network to messages (To make it out of sync)
		node3.Communication.DisableIncomingCommunication()

		// node0, node1, and node2 are closing block1
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)
		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		net.WaitUntilNodesEventuallyCommitASpecificBlock(ctx, t, 0, block1, node0, node1, node2)

		// node3 is still "stuck" on the genesis block
		node3LatestBlock := node3.GetLatestBlock()
		require.True(t, node3LatestBlock == interfaces.GenesisBlock, "node3 should have been on genesis but its latest block is %s", node3LatestBlock)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // hangs on block2

		bc, err := leaderelection.GenerateBlocksWithProofsForTest([]interfaces.Block{block1, block2, block3}, net.Nodes)
		if err != nil {
			t.Fatalf("Error creating mock blockchain for tests - %s", err)
			return
		}
		blockToSync, blockProofToSync := bc.BlockAndProofAt(2)
		prevBlockProofToSync := bc.BlockProofAt(1)
		if err := node3.Sync(ctx, blockToSync, blockProofToSync, prevBlockProofToSync); err != nil {
			t.Fatalf("Sync failed for node %s - %s", node3.MemberId, err)
		}
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 3, node3)
		require.True(t, node3.GetCurrentHeight() >= block2.Height())

		// opening node3's network to messages
		node3.Communication.EnableIncomingCommunication()

		net.SetNodesToNotPauseOnRequestNewBlock()
		net.ResumeRequestNewBlockOnNodes(ctx, node0)

		// Just verify all nodes reached the height of block2, either by commit or by sync.
		// DON'T try to verify the next block is committed (block3) because sometimes one of the nodes gets left behind
		// Normally that node would trigger a sync and get the block and continue closing blocks with the other nodes,
		// but during this test there is no additional sync so the node will never catch up, thus it will never commit block3
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, block2.Height(), node0, node1, node2, node3)
	})
}
