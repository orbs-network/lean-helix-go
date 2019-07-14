// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leaderelection

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

const LOG_TO_CONSOLE = false

func Test2fPlus1ViewChangeToBeElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		h := NewHarness(ctx, t, LOG_TO_CONSOLE, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		// hang the leader (node0)
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0)
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{})

		// manually cause new-view with 3 view-changes
		node0VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node0.KeyManager, node0.MemberId, 1, 1, nil)
		node2VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil)
		node3VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node3.KeyManager, node3.MemberId, 1, 1, nil)
		node1.Communication.OnRemoteMessage(ctx, node0VCMessage.ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, node2VCMessage.ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, node3VCMessage.ToConsensusRawMessage())

		// now that we caused node1 to be the new leader, he'll ask for a new block (block2)
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node1)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node1)

		// release the hanged the leader (node0)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)

		// make sure that we're on block2
		h.net.WaitForAllNodesToCommitBlock(ctx, block2)
	})
}

func TestBlockIsNotUsedWhenElectionHappened(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block1)

		h := NewHarness(ctx, t, true, block1, block2, block3)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]

		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		h.net.WaitForAllNodesToCommitBlock(ctx, block1)

		// Thwart Preprepare message sending by node0 for block2
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0)           // pause when proposing block2

		// increment view - this selects node1 as the leader
		/*
			All nodes progress to the next view:
			we blocked PREPREPARE from being sent by the leader node0
			so other nodes did not receive it and send out PREPARE
			so in turn they did not receive 2f+1 PREPAREs (a.k.a PREPARED phase)
			so new leader is free to suggest another block instead of block2
		*/

		<- h.net.Nodes[1].TriggerElectionOnNode(ctx)
		<- h.net.Nodes[2].TriggerElectionOnNode(ctx)
		<- h.net.Nodes[3].TriggerElectionOnNode(ctx)

		// free the first leader to send stale PREPREPARE now when the others are in next view
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)

		// tell the old leader to advance it's view so it can join the others in view 1
		<- h.net.Nodes[0].TriggerElectionOnNode(ctx)

		// sync with new leader on block proposal
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node1)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node1) // processing block 3
		fmt.Printf("Resumed RequestNewBlock on node1\n")
		h.net.WaitForAllNodesToCommitBlock(ctx, block2)

		// TODO - expect preprepare messages were sent from node0 for block2
	})
}

/*
func TestBlockIsNotUsedWhenElectionHappened(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block1)

		h := NewHarness(ctx, t, block1, block2, block3)

		node0 := h.net.Nodes[0]
		//node1 := h.net.Nodes[1]

		// processing block1, should be agreed by all nodes

		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)

		h.net.WaitForAllNodesToCommitBlock(ctx, block1)

		// processing block 2
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0)

		node0.Communication.DisableOutgoing() // Rationale: do not send PREPREPARE, prevent race-condition

		// select node 1 as the leader (dropping block2)
		h.net.Nodes[0].TriggerElection(ctx)
		h.net.Nodes[1].TriggerElection(ctx)
		h.net.Nodes[2].TriggerElection(ctx)
		h.net.Nodes[3].TriggerElection(ctx)

		node0.Communication.EnableOutgoing()


		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		//h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node1)

		// processing block 3
		//h.net.ResumeRequestNewBlockOnNodes(ctx, node1)
		h.net.WaitForAllNodesToCommitBlock(ctx, block2)
	})
}
*/
func TestThatNewLeaderSendsNewViewWhenElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t, LOG_TO_CONSOLE)
		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0)
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{})

		// selection node 1 as the leader
		node0.TriggerElectionOnNode(ctx)
		node1.TriggerElectionOnNode(ctx)
		node2.TriggerElectionOnNode(ctx)
		node3.TriggerElectionOnNode(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node1)
		node0.Communication.ClearOutgoingWhitelist()

		h.net.ResumeRequestNewBlockOnNodes(ctx, node1)
		h.net.WaitForAllNodesToCommitTheSameBlock(ctx)

		require.Equal(t, 1, node0.Communication.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node2.Communication.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node3.Communication.CountSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE))
		require.Equal(t, 1, node1.Communication.CountSentMessages(protocol.LEAN_HELIX_NEW_VIEW))
	})
}

func TestNotCountingViewChangeFromTheSameNode(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		h := NewHarness(ctx, t, LOG_TO_CONSOLE, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]

		// hang the leader (node0)
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0)

		// sending only 4 view-change from the same node
		node1.Communication.OnRemoteMessage(ctx, builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil).ToConsensusRawMessage())

		node1.Communication.CountSentMessages(protocol.LEAN_HELIX_NEW_VIEW)
	})
}

func TestNoNewViewIfLessThan2fPlus1ViewChange(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		h := NewHarness(ctx, t, LOG_TO_CONSOLE, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]

		// hang the leader (node0)
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, node0)

		// sending only 2 view-change (not enough to be elected)
		node0VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node0.KeyManager, node0.MemberId, 1, 1, nil)
		node2VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil)
		node1.Communication.OnRemoteMessage(ctx, node0VCMessage.ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, node2VCMessage.ToConsensusRawMessage())

		// release the hanged the leader (node0)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)

		// make sure that we're on block2
		h.net.WaitForAllNodesToCommitBlock(ctx, block2)
	})
}

// TODO: This is sometimes stuck!!! Remove this comment if doesnt happen by end of June 2019
func TestLeaderCircularOrdering(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		// Nodes might get into prepared state, and send their block in the view-change
		// meaning that the new leader will not request new block and we can't hang him.
		// to prevent nodes from getting prepared, we just don't validate the block
		h := NewHarnessWithFailingBlockProposalValidations(ctx, t, LOG_TO_CONSOLE)

		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, h.net.Nodes[0])

		// selecting node 1 as the leader
		h.net.Nodes[0].TriggerElectionOnNode(ctx)
		h.net.Nodes[2].TriggerElectionOnNode(ctx)
		h.net.Nodes[3].TriggerElectionOnNode(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[0])
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, h.net.Nodes[1])

		// selecting node 2 as the leader
		h.net.Nodes[0].TriggerElectionOnNode(ctx)
		h.net.Nodes[1].TriggerElectionOnNode(ctx)
		h.net.Nodes[3].TriggerElectionOnNode(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[1])
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, h.net.Nodes[2])

		// selecting node 3 as the leader
		h.net.Nodes[0].TriggerElectionOnNode(ctx)
		h.net.Nodes[1].TriggerElectionOnNode(ctx)
		h.net.Nodes[2].TriggerElectionOnNode(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[2])
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, h.net.Nodes[3])

		// back to node 0 as the leader
		h.net.Nodes[1].TriggerElectionOnNode(ctx)
		h.net.Nodes[2].TriggerElectionOnNode(ctx)
		h.net.Nodes[3].TriggerElectionOnNode(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[3])
		h.net.ReturnWhenNodePausesOnRequestNewBlock(ctx, h.net.Nodes[0])
	})
}
