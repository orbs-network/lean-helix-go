// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHappyFlow(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		net := network.ABasicTestNetworkWithTimeBasedElectionsAndConsoleLogs(ctx, t)
		net.StartConsensus(ctx)
		net.WaitUntilQuorumOfNodesEventuallyReachASpecificHeight(ctx, 1)
	})
}

func TestHappyFlowMessages(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		net := network.ABasicTestNetwork(ctx)
		net.SetNodesToPauseOnRequestNewBlock()

		net.StartConsensus(ctx)

		// let the leader run on the first round
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, net.Nodes[0])
		net.ResumeRequestNewBlockOnNodes(ctx, net.Nodes[0])

		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2)

		// hang the leader before the next round
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, net.Nodes[0])

		require.Equal(t, 1, net.Nodes[0].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[1].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[2].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))
		require.Equal(t, 0, net.Nodes[3].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPREPARE))

		require.Equal(t, 0, net.Nodes[0].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[1].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[2].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))
		require.Equal(t, 1, net.Nodes[3].Communication.CountSentMessages(protocol.LEAN_HELIX_PREPARE))

		require.Equal(t, 1, net.Nodes[0].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[1].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[2].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
		require.Equal(t, 1, net.Nodes[3].Communication.CountSentMessages(protocol.LEAN_HELIX_COMMIT))
	})
}

func TestConsensusFor2Blocks(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		net := network.
			ABasicTestNetworkWithTimeBasedElectionsAndConsoleLogs(ctx, t).
			StartConsensus(ctx)
		net.WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx, 2, 1) // We must not wait for ALL nodes to commit H=8 as sometimes one of the nodes will start after PREPREPARE is sent so it will be left behind for good (there is no NodeSync to save it as in production case)
	})
}

func TestHangingNode(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		net := network.ATestNetworkBuilder(4, block1, block2).
			//LogToConsole(t).
			Build(ctx)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		net.SetNodesToPauseOnValidateBlock()
		net.StartConsensus(ctx)
		// TODO This hangs, maybe impl is bad, compare to RequestNewBlockProposal
		net.ReturnWhenNodesPauseOnValidateBlock(ctx, node1, node2, node3)
		net.ResumeValidateBlockOnNodes(ctx, node1, node2)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2, node0, node1, node2)

		node0LatestBlock := node0.GetLatestBlock()
		node1LatestBlock := node1.GetLatestBlock()
		node2LatestBlock := node2.GetLatestBlock()
		node3LatestBlock := node3.GetLatestBlock()

		if node0LatestBlock == nil {
			fmt.Printf("Weird: node0 latest is nil!")
			t.Fatal("node0Latest is nil")
		}
		require.True(t, matchers.BlocksAreEqual(node0LatestBlock, block1), "%s should be equal to %s", node0LatestBlock, block1)
		require.True(t, matchers.BlocksAreEqual(node1LatestBlock, block1), "%s should be equal to %s", node1LatestBlock, block1)
		require.True(t, matchers.BlocksAreEqual(node2LatestBlock, block1), "%s should be equal to %s", node2LatestBlock, block1)
		require.True(t, node3LatestBlock == interfaces.GenesisBlock)

		net.ReturnWhenNodesPauseOnValidateBlock(ctx, node1, node2)
		net.ResumeValidateBlockOnNodes(ctx, node1, node2)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 3, node0, node1, node2)
		node0LatestBlock = node0.GetLatestBlock()
		node1LatestBlock = node1.GetLatestBlock()
		node2LatestBlock = node2.GetLatestBlock()
		node3LatestBlock = node3.GetLatestBlock()

		require.True(t, matchers.BlocksAreEqual(node0LatestBlock, block2), "%s should be equal to %s", node0LatestBlock, block2)
		require.True(t, matchers.BlocksAreEqual(node1LatestBlock, block2), "%s should be equal to %s", node1LatestBlock, block2)
		require.True(t, matchers.BlocksAreEqual(node2LatestBlock, block2), "%s should be equal to %s", node2LatestBlock, block2)
		require.True(t, node3LatestBlock == interfaces.GenesisBlock)

		net.ResumeValidateBlockOnNodes(ctx, node3)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2, node3)
		node3LatestBlock = node3.GetLatestBlock()
		require.True(t, matchers.BlocksAreEqual(node3LatestBlock, block1), "%s should be equal to %s", node3LatestBlock, block1)

		net.ReturnWhenNodesPauseOnValidateBlock(ctx, node3)
		net.ResumeValidateBlockOnNodes(ctx, node3)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 3, node3)
		node3LatestBlock = node3.GetLatestBlock()
		require.True(t, matchers.BlocksAreEqual(node3LatestBlock, block2), "%s should be equal to %s", node3LatestBlock, block2)
	})
}
