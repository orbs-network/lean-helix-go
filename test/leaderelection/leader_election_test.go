// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const LOG_TO_CONSOLE = false

func TestNewLeaderProposesNewBlock_IfPreviousLeaderFailedToBringNetworkIntoPreparedPhase(t *testing.T) {
	test.WithContextWithTimeout(t, 15*time.Second, func(ctx context.Context) {
		h := NewStartedHarness(ctx, t, LOG_TO_CONSOLE)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)
		node0.Communication.DisableOutgoingCommunication()

		h.net.TriggerElectionsOnAllNodes(ctx)

		// Now that we caused node1 to be the new leader, it will ask for a new block.
		// BTW the test doesn't care which block it actually is
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
	})
}

func TestNotCountingViewChangeFromTheSameNode(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewStartedHarness(ctx, t, LOG_TO_CONSOLE)
		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

		// sending only 4 view-change from the same node
		for i := 0; i < 4; i++ {
			node1.Communication.OnIncomingMessage(ctx, builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())
		}
		require.Zero(t, node1.Communication.CountSentMessages(protocol.LEAN_HELIX_NEW_VIEW), "node1 sent new view although it didn't receive enough valid votes")
	})
}

func TestDoesNotReachConsensusOnBlockWhenValidateBlockProposalFails(t *testing.T) {
	test.WithContext(func(ctx context.Context) {

		h := NewStartedHarnessWithFailingBlockProposalValidations(ctx, t, LOG_TO_CONSOLE)

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, h.net.Nodes[0])
		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[0])

		c := make(chan struct{})
		go func() {
			h.net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2)
			close(c)
		}()

		select {
		case <-time.After(50 * time.Millisecond):
		case <-c:
			t.Fatal("Reached consensus on block despite validations failing")
		}
	})
}

func TestBlockIsNotUsedWhenElectionHappened(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block1)

		h := NewStartedHarness(ctx, t, LOG_TO_CONSOLE, block1, block2, block3)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		h.net.WaitUntilNodesEventuallyCommitASpecificBlock(ctx, t, 0, block1)

		// Thwart Preprepare message sending by node0 for block2
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // pause when proposing block2

		// increment view - this selects node1 as the leader
		/*
			All nodes progress to the next view:
			we blocked PREPREPARE from being sent by the leader node0
			so other nodes did not receive it and send out PREPARE
			so in turn they did not receive 2f+1 PREPAREs (a.k.a PREPARED phase)
			so new leader is free to suggest another block instead of block2
		*/

		<-h.net.Nodes[1].TriggerElectionOnNode(ctx)
		<-h.net.Nodes[2].TriggerElectionOnNode(ctx)
		<-h.net.Nodes[3].TriggerElectionOnNode(ctx)

		// free the first leader to send stale PREPREPARE now when the others are in next view
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		// tell the old leader to advance it's view so it can join the others in view 1
		<-h.net.Nodes[0].TriggerElectionOnNode(ctx)
		// sync with new leader on block proposal
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node1) // processing block 3
		h.net.WaitUntilNodesEventuallyCommitASpecificBlock(ctx, t, 0, block3)
	})
}

func TestThatNewLeaderSendsNewViewWhenElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewStartedHarness(ctx, t, LOG_TO_CONSOLE)
		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)
		node0.Communication.DisableOutgoingCommunication()

		h.net.WaitUntilNetworkIsRunning(ctx)

		// will elect node1 as the leader (because nodes are elected sequentially)
		h.net.TriggerElectionsOnAllNodes(ctx)

		//h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
		node0.Communication.EnableOutgoingCommunication()

		h.net.ResumeRequestNewBlockOnNodes(ctx, node1)
		h.net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 2)

		require.Equal(t, 1, node0.Communication.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE)) // node0's send of view change will be counted even though its comms are down, since we count *attempts* to send messages rather than *successful* sends
		require.Equal(t, 1, node2.Communication.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node3.Communication.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node1.Communication.CountSentMessages(protocol.LEAN_HELIX_NEW_VIEW))
	})
}

func TestViewNotIncrementedIfLessThan2fPlus1ViewChange(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		h := NewStartedHarness(ctx, t, LOG_TO_CONSOLE, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]

		// Verify leader (node0) indeed starts RequestNewBlockProposal()
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

		// sending only 2 VIEW_CHANGE
		// This is not enough to be elected as f=1 for 4 nodes, so 2f+1 is 3 nodes
		node0VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node0.KeyManager, node0.MemberId, 1, 1, nil)
		node2VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil)
		node1.Communication.OnIncomingMessage(ctx, node0VCMessage.ToConsensusRawMessage())
		node1.Communication.OnIncomingMessage(ctx, node2VCMessage.ToConsensusRawMessage())

		// Resume the paused leader (node0)
		//h.net.ResumeRequestNewBlockOnNodes(ctx, node0)

		h.net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 1, node0)
		h.net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 1, node1)

		go func() {
			// Fail if node1 starts RequestNewBlockProposal() because it means it became new leader
			h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
			if ctx.Err() == nil {
				t.Fatal("node1 tried to propose a block after receiving only 2 view change messages")
			}
		}()

		time.Sleep(100 * time.Millisecond)
		// node 1 got a chance to propose a block and did not take it as expected
	})
}
