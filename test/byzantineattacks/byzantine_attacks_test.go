// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package byzantineattacks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestThatWeReachConsensusWhere1OutOf4NodeIsByzantine(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)
		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]interfaces.Block{block}).
			Build(ctx)

		net.Nodes[3].Communication.SetIncomingWhitelist([]primitives.MemberId{})

		net.StartConsensus(ctx)

		net.WaitForNodesToCommitABlock(ctx, net.Nodes[0], net.Nodes[1], net.Nodes[2])
	})
}

func TestThatWeReachConsensusWhere2OutOf7NodesAreByzantine(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)
		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(7).
			WithBlocks([]interfaces.Block{block}).
			Build(ctx)

		net.Nodes[1].Communication.SetIncomingWhitelist([]primitives.MemberId{})
		net.Nodes[2].Communication.SetIncomingWhitelist([]primitives.MemberId{})

		net.StartConsensus(ctx)

		net.WaitForNodesToCommitABlock(ctx, net.Nodes[0], net.Nodes[3], net.Nodes[4], net.Nodes[5], net.Nodes[6])
	})
}

func TestThatAByzantineLeaderCanNotCauseAForkBySendingTwoBlocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]interfaces.Block{block1}).
			Build(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		//node3 := net.Nodes[3]

		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{node1.MemberId, node2.MemberId})

		// the leader (node0) is suggesting block1 to node1 and node2 (not to node3)
		net.StartConsensus(ctx)

		// node0, node1 and node2 should reach consensus
		net.WaitForNodesToCommitASpecificBlock(ctx, block1, node0, node1, node2)

		node0.StartConsensus(ctx)
	})
}

func TestNoForkWhenAByzantineNodeSendsABadBlockSeveralTimes(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		goodBlock := mocks.ABlock(interfaces.GenesisBlock)
		fakeBlock := mocks.ABlock(interfaces.GenesisBlock)
		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]interfaces.Block{goodBlock}).
			LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		byzantineNode := net.Nodes[3]
		net.SetNodesToPauseOnRequestNewBlock(node0)
		net.StartConsensus(ctx)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

		// fake a preprepare message from node3 (byzantineNode) that points to a unrelated block (Should be ignored)
		ppm := builders.APreprepareMessage(net.InstanceId, byzantineNode.KeyManager, byzantineNode.MemberId, 1, 1, fakeBlock)
		byzantineNode.Communication.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())
		byzantineNode.Communication.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())
		byzantineNode.Communication.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())
		byzantineNode.Communication.SendConsensusMessage(ctx, []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}, ppm.ToConsensusRawMessage())

		net.ResumeRequestNewBlockOnNodes(ctx, node0)

		require.True(t, net.WaitForAllNodesToCommitBlockAndReturnWhetherEqualToGiven(ctx, goodBlock))
	})
}

func TestThatAByzantineLeaderCannotCauseAFork(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(interfaces.GenesisBlock)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]interfaces.Block{block1, block2}).
			LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{node1.MemberId, node2.MemberId})
		node1.Communication.SetOutgoingWhitelist([]primitives.MemberId{node2.MemberId})
		node2.Communication.SetOutgoingWhitelist([]primitives.MemberId{})
		node3.Communication.SetOutgoingWhitelist([]primitives.MemberId{})

		net.StartConsensus(ctx)

		// Because we only allow node0 (The leader) to talk to node1 and node2
		// and node1 only to talk to node2,
		// we can expect (only) node2 to be prepared on block1
		test.Eventually(time.Duration(100)*time.Millisecond, func() bool {
			_, ppOk := node2.Storage.GetPreprepareMessage(1, 1)
			p, _ := node2.Storage.GetPrepareMessages(1, 1, mocks.CalculateBlockHash(block1))
			return ppOk && len(p) == 2
		})

		// now that node2 is prepared on block1, we'll close any communication
		// to it, and open all the other nodes communication.
		// then, we trigger an election. Node2's prepared block will not get sent in a view-change
		node0.Communication.ClearOutgoingWhitelist()
		node1.Communication.ClearOutgoingWhitelist()
		node2.Communication.ClearOutgoingWhitelist()
		node3.Communication.ClearOutgoingWhitelist()

		node2.Communication.SetOutgoingWhitelist([]primitives.MemberId{})
		node2.Communication.SetIncomingWhitelist([]primitives.MemberId{})

		// selection node 1 as the leader
		node0.TriggerElectionOnNode(ctx)
		node1.TriggerElectionOnNode(ctx)
		node2.TriggerElectionOnNode(ctx)
		node3.TriggerElectionOnNode(ctx)

		net.WaitForNodesToCommitASpecificBlock(ctx, block2, node0, node1, node3)
	})
}
