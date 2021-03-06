// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package byzantineattacks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/orbs-network/lean-helix-go/testhelpers"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestThatWeReachConsensusWhere1of4NodeIsByzantine(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		block := mocks.ABlock(interfaces.GenesisBlock)
		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks(block).
			WithTimeBasedElectionTrigger(1000 * time.Millisecond).
			LogToConsole(t).
			Build(ctx)

		net.Nodes[3].Communication.DisableIncomingCommunication()

		net.StartConsensus(ctx)

		net.WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx, 2, 1)
	})
}

func TestNetworkReachesConsensusWhen2of7NodesAreByzantine(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {

		//block := mocks.ABlock(interfaces.GenesisBlock)
		totalNodes := 7
		honestNodes := quorum.CalcQuorumWeight(testhelpers.EvenWeights(totalNodes))
		net := network.
			NewTestNetworkBuilder().
			LogToConsole(t).
			WithNodeCount(totalNodes).
			WithTimeBasedElectionTrigger(1000 * time.Millisecond). // reducing the timeout is flaky since sync is not performed and nodes may drop out if interrupted too frequently
			//WithBlocks(block).
			Build(ctx)

		byzantineNodes := net.Nodes[honestNodes:totalNodes]
		for _, b := range byzantineNodes {
			b.Communication.DisableIncomingCommunication()
		}

		net.StartConsensus(ctx)

		net.WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx, 2, 1)
	})
}

// TODO Flaky
// TODO This is a weak test, it only tests that 3 nodes out of 4 can close a block.
// It does not test what happens if the leader sends block1a to node1,node2 and block1b to node3
// where block1a and block1b both have height=1 but different contents.
func TestThatAByzantineLeaderCanNotCauseAForkBySendingTwoBlocks(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithTimeBasedElectionTrigger(1000 * time.Millisecond).
			WithBlocks(block1).
			Build(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]

		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{
			node1.MemberId,
			node2.MemberId,
		})

		// the leader (node0) is suggesting block1 to node1 and node2 (not to node3)
		net.StartConsensus(ctx)

		// node0, node1 and node2 should reach consensus
		net.WaitUntilNodesEventuallyCommitASpecificBlock(ctx, t, 0, block1, node0, node1, node2)
	})
}

func TestNoForkWhenAByzantineNodeSendsABadBlockSeveralTimes(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		goodBlock := mocks.ABlock(interfaces.GenesisBlock)
		fakeBlock := mocks.ABlock(interfaces.GenesisBlock)
		t.Logf("GoodBlock=%s FakeBlock=%s", goodBlock, fakeBlock)

		require.False(t, matchers.BlocksAreEqual(goodBlock, fakeBlock))

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithTimeBasedElectionTrigger(1000 * time.Millisecond).
			WithBlocks(goodBlock).
			//LogToConsole().
			Build(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		honestNodes := []primitives.MemberId{node0.MemberId, node1.MemberId, node2.MemberId}
		byzantineNode := net.Nodes[3]

		net.SetNodesToPauseOnRequestNewBlock(node0)
		net.StartConsensus(ctx)

		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

		// fake a preprepare message from node3 (byzantineNode) that points to a unrelated block (Should be ignored)
		ppm := builders.APreprepareMessage(net.InstanceId, byzantineNode.KeyManager, byzantineNode.MemberId, 1, 1, fakeBlock).ToConsensusRawMessage()
		_ = byzantineNode.Communication.SendConsensusMessage(ctx, honestNodes, ppm)
		_ = byzantineNode.Communication.SendConsensusMessage(ctx, honestNodes, ppm)
		_ = byzantineNode.Communication.SendConsensusMessage(ctx, honestNodes, ppm)
		_ = byzantineNode.Communication.SendConsensusMessage(ctx, honestNodes, ppm)

		net.ResumeRequestNewBlockOnNodes(ctx, node0)
		t.Logf("Waiting for commit of good block %s", goodBlock)

		net.WaitUntilNodesEventuallyCommitASpecificBlock(ctx, t, 0, goodBlock)
	})
}

func TestThatAByzantineLeaderCannotCauseAFork(t *testing.T) {
	t.Skip("This purpose of this test needs to be clarified, it must be rewritten and become shorter than it is now")
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(interfaces.GenesisBlock)

		net := network.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks(block1, block2).
			//LogToConsole(t).
			Build(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{node1.MemberId, node2.MemberId})
		node1.Communication.SetOutgoingWhitelist([]primitives.MemberId{node2.MemberId})
		node2.Communication.DisableOutgoingCommunication()
		node3.Communication.DisableOutgoingCommunication()

		net.StartConsensus(ctx)

		// Because we only allow node0 (The leader) to talk to node1 and node2
		// and node1 only to talk to node2,
		// we can expect (only) node2 to be prepared on block1
		test.Eventually(100*time.Millisecond, func() bool {
			_, ppOk := node2.Storage.GetPreprepareMessage(1, 1)
			p, _ := node2.Storage.GetPrepareMessages(1, 1, mocks.CalculateBlockHash(block1))
			return ppOk && len(p) == 2
		})

		// now that node2 is prepared on block1, we'll close any communication
		// to it, and open all the other nodes communication.
		// then, we trigger an election. Node2's prepared block will not get sent in a view-change
		node0.Communication.EnableOutgoingCommunication()
		node1.Communication.EnableOutgoingCommunication()
		node2.Communication.EnableOutgoingCommunication()
		node3.Communication.EnableOutgoingCommunication()

		node2.Communication.DisableOutgoingCommunication()
		node2.Communication.DisableIncomingCommunication()

		// selection node 1 as the leader
		node0.TriggerElectionOnNode(ctx)
		node1.TriggerElectionOnNode(ctx)
		node2.TriggerElectionOnNode(ctx)
		node3.TriggerElectionOnNode(ctx)

		net.WaitUntilNodesEventuallyCommitASpecificBlock(ctx, t, 0, block2, node0, node1, node3)
	})
}
