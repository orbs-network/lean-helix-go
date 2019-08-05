// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHappyFlow(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetworkWithConsoleLogs(ctx, t)
		net.StartConsensus(ctx)
		net.WaitUntilQuorumCommitsHeight(ctx, 1)
		//net.WaitUntilNodesCommitASpecificHeight(ctx, 1)
	})
}

func TestHappyFlowMessages(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork(ctx)
		net.SetNodesToPauseOnRequestNewBlock()

		net.StartConsensus(ctx)

		// let the leader run on the first round
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, net.Nodes[0])
		net.ResumeRequestNewBlockOnNodes(ctx, net.Nodes[0])

		net.MAYBE_FLAKY_WaitForAllNodesToCommitTheSameBlock(ctx)

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

func TestConsensusFor8Blocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetworkWithConsoleLogs(ctx, t).StartConsensus(ctx)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 8)
	})
}

// TODO Flaky
func TestHangingNode(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		net := network.ATestNetworkBuilder(4, block1, block2).Build(ctx)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		net.SetNodesToPauseOnValidateBlock()
		net.StartConsensus(ctx)

		net.ReturnWhenNodesPauseOnValidateBlock(ctx, node1, node2, node3)
		net.ResumeValidateBlockOnNodes(ctx, node1, node2)
		net.WaitUntilNodesCommitAnyBlock(ctx, node0, node1, node2)
		require.True(t, matchers.BlocksAreEqual(node0.GetLatestBlock(), block1))
		require.True(t, matchers.BlocksAreEqual(node1.GetLatestBlock(), block1))
		require.True(t, matchers.BlocksAreEqual(node2.GetLatestBlock(), block1))
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.ReturnWhenNodesPauseOnValidateBlock(ctx, node1, node2)
		net.ResumeValidateBlockOnNodes(ctx, node1, node2)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 3, node0, node1, node2)
		require.True(t, matchers.BlocksAreEqual(node0.GetLatestBlock(), block2))
		require.True(t, matchers.BlocksAreEqual(node1.GetLatestBlock(), block2))
		require.True(t, matchers.BlocksAreEqual(node2.GetLatestBlock(), block2))
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.ResumeValidateBlockOnNodes(ctx, node3)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2, node3)
		require.True(t, matchers.BlocksAreEqual(node3.GetLatestBlock(), block1))

		net.ReturnWhenNodesPauseOnValidateBlock(ctx, node3)
		net.ResumeValidateBlockOnNodes(ctx, node3)
		net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 3, node3)
		require.True(t, matchers.BlocksAreEqual(node3.GetLatestBlock(), block2))
	})
}
