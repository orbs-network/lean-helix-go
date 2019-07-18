// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const LOG_TO_CONSOLE = true

func TestNewLeaderProposesNewBlockIfPreviousLeaderFailedToBringNetworkIntoPreparedPhase(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		h := NewHarness(ctx, t, LOG_TO_CONSOLE, block1, block2)
		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

		node0.Communication.DisableOutgoingCommunication()

		manuallyElectNode1AsNewLeader(ctx, h)

		// Now that we caused node1 to be the new leader, it will ask for a new block.
		// BTW the test doesn't care which block it actually is
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
	})
}

func manuallyElectNode1AsNewLeader(ctx context.Context, h *harness) {
	node0 := h.net.Nodes[0]
	node1 := h.net.Nodes[1]
	node2 := h.net.Nodes[2]
	node3 := h.net.Nodes[3]

	node0VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node0.KeyManager, node0.MemberId, 1, 1, nil)
	node2VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil)
	node3VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node3.KeyManager, node3.MemberId, 1, 1, nil)
	node1.Communication.OnRemoteMessage(ctx, node0VCMessage.ToConsensusRawMessage())
	node1.Communication.OnRemoteMessage(ctx, node2VCMessage.ToConsensusRawMessage())
	node1.Communication.OnRemoteMessage(ctx, node3VCMessage.ToConsensusRawMessage())
}

func TestBlockIsNotUsedWhenElectionHappened(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block1)

		h := NewHarness(ctx, t, LOG_TO_CONSOLE, block1, block2, block3)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // processing block1, should be agreed by all nodes
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		require.True(t, h.net.WaitForAllNodesToCommitBlockAndReturnWhetherEqualToGiven(ctx, block1))
		t.Log("--- BLOCK1 COMMITTED ---")
		// Thwart Preprepare message sending by node0 for block2
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0) // pause when proposing block2
		t.Log("--- NODE0 PAUSED ON REQUEST NEW BLOCK ---")

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

		t.Log("--- TRIGGERED ELECTION ON NODES 1 2 3 ---")

		// free the first leader to send stale PREPREPARE now when the others are in next view
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		t.Log("--- NODE0 RESUMED REQUEST NEW BLOCK ---")
		// tell the old leader to advance it's view so it can join the others in view 1
		<-h.net.Nodes[0].TriggerElectionOnNode(ctx)
		t.Log("--- TRIGGERED ELECTION ON NODE0")
		// sync with new leader on block proposal
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
		t.Log("--- NODE1 PAUSED ON REQUEST NEW BLOCK ---")
		h.net.ResumeRequestNewBlockOnNodes(ctx, node1) // processing block 3
		t.Log("--- NODE1 RESUMED REQUEST NEW BLOCK ---")
		require.True(t, h.net.WaitForAllNodesToCommitBlockAndReturnWhetherEqualToGiven(ctx, block3))
	})
}

func TestThatNewLeaderSendsNewViewWhenElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t, LOG_TO_CONSOLE)
		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{})

		// Elect node1 as the leader
		node0.TriggerElectionOnNode(ctx)
		node1.TriggerElectionOnNode(ctx)
		node2.TriggerElectionOnNode(ctx)
		node3.TriggerElectionOnNode(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
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
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

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
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)

		// sending only 2 view-change (not enough to be elected)
		node0VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node0.KeyManager, node0.MemberId, 1, 1, nil)
		node2VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil)
		node1.Communication.OnRemoteMessage(ctx, node0VCMessage.ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, node2VCMessage.ToConsensusRawMessage())

		// Resume the paused leader (node0)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)

		// Make sure we're on block1
		require.True(t, h.net.WaitForAllNodesToCommitBlockAndReturnWhetherEqualToGiven(ctx, block1))

		node1TriesToProposeABlock := make(chan struct{})
		go func() {
			h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node1)
			node1TriesToProposeABlock <- struct{}{}
		}()

		shortCtx, shortCancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer shortCancel()
		select {
		case <-shortCtx.Done():
			t.Log("node 1 got a chance to propose a block and did not take it as expected")
		case <-node1TriesToProposeABlock:
			t.Fatal("node1 tried to propose a block after receiving only 2 view change messages")
		}
	})
}

// Let each and every node try and be the Leader and finally return to the original leader
func TestLeaderCircularOrdering(t *testing.T) {
	test.WithContext(func(ctx context.Context) {

		// TL;DR Always fail validation so that the network will never close blocks

		// Set block validation to always fail.
		// The reason for this is to prevent the Validator (non-leader) nodes
		// from going into PREPARED phase after validating the block.
		// If nodes were to go into PREPARED phase, this would "lock" the proposed
		// block, preventing the next Leader from suggesting a different block
		// by calling RequestNewBlockProposal.
		// We DO want node0 to pause on RequestNewBlockProposal because it is our stop signal for the test

		timer := time.AfterFunc(50000*time.Millisecond, func() {
			t.Fatal("Test is stuck")
		})
		h := NewHarnessWithFailingBlockProposalValidations(ctx, t, LOG_TO_CONSOLE)

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, h.net.Nodes[0])
		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[0])

		electNewLeader(ctx, h, 1)
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, h.net.Nodes[1])
		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[1])

		electNewLeader(ctx, h, 2)
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, h.net.Nodes[2])
		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[2])

		electNewLeader(ctx, h, 3)
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, h.net.Nodes[3])
		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[3])

		// back to node0 as leader
		electNewLeader(ctx, h, 0)
		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, h.net.Nodes[0])
		timer.Stop()
	})
}

func electNewLeader(ctx context.Context, h *harness, newLeaderIndex int) {

	for i, node := range h.net.Nodes {
		if i == newLeaderIndex {
			continue
		}
		<-node.TriggerElectionOnNode(ctx)
	}
}

func TestDoesNotCloseBlockWhenValidateBlockProposalFails(t *testing.T) {
	test.WithContext(func(ctx context.Context) {

		h := NewHarnessWithFailingBlockProposalValidations(ctx, t, LOG_TO_CONSOLE)

		h.net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, h.net.Nodes[0])
		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[0])

		c := make(chan struct{})
		go func() {
			h.net.WaitForNodesToCommitABlock(ctx)
			close(c)
		}()

		select {
		case <-time.After(50 * time.Millisecond):
		case <-c:
			t.Fatal("Block was closed despite validations failing")
		}
	})
}
