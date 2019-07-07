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
)

func Test2fPlus1ViewChangeToBeElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)

		h := NewHarness(ctx, t, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		// hang the leader (node0)
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{})

		// manually cause new-view with 3 view-changes
		node0VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node0.KeyManager, node0.MemberId, 1, 1, nil)
		node2VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node2.KeyManager, node2.MemberId, 1, 1, nil)
		node3VCMessage := builders.AViewChangeMessage(h.net.InstanceId, node3.KeyManager, node3.MemberId, 1, 1, nil)
		node1.Communication.OnRemoteMessage(ctx, node0VCMessage.ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, node2VCMessage.ToConsensusRawMessage())
		node1.Communication.OnRemoteMessage(ctx, node3VCMessage.ToConsensusRawMessage())

		// now that we caused node1 to be the new leader, he'll ask for a new block (block2)
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node1)
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

		h := NewHarness(ctx, t, block1, block2, block3)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]

		// processing block1, should be agreed by all nodes
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		h.net.WaitForAllNodesToCommitBlock(ctx, block1)

		// processing block 2
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{})
		// selection node 1 as the leader (dropping block2)
		h.net.Nodes[0].TriggerElection(ctx)
		h.net.Nodes[1].TriggerElection(ctx)
		h.net.Nodes[2].TriggerElection(ctx)
		h.net.Nodes[3].TriggerElection(ctx)

		node0.Communication.ClearOutgoingWhitelist()

		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node1)
		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)

		// processing block 3
		h.net.ResumeRequestNewBlockOnNodes(ctx, node1)
		h.net.WaitForAllNodesToCommitBlock(ctx, block3)
	})
}

func TestThatNewLeaderSendsNewViewWhenElected(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)
		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]
		node3 := h.net.Nodes[3]

		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)
		node0.Communication.SetOutgoingWhitelist([]primitives.MemberId{})

		// selection node 1 as the leader
		node0.TriggerElection(ctx)
		node1.TriggerElection(ctx)
		node2.TriggerElection(ctx)
		node3.TriggerElection(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, node0)
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node1)
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

		h := NewHarness(ctx, t, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]

		// hang the leader (node0)
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)

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

		h := NewHarness(ctx, t, block1, block2)

		node0 := h.net.Nodes[0]
		node1 := h.net.Nodes[1]
		node2 := h.net.Nodes[2]

		// hang the leader (node0)
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)

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
		h := NewHarness(ctx, t)

		// Nodes might get into prepared state, and send their block in the view-change
		// meaning that the new leader will not request new block and we can't hang him.
		// to prevent nodes from getting prepared, we just don't validate the block

		h.net.Nodes[0].BlockUtils.SetValidationResult(false)
		h.net.Nodes[1].BlockUtils.SetValidationResult(false)
		h.net.Nodes[2].BlockUtils.SetValidationResult(false)
		h.net.Nodes[3].BlockUtils.SetValidationResult(false)

		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, h.net.Nodes[0])

		// selecting node 1 as the leader
		h.net.Nodes[0].TriggerElection(ctx)
		h.net.Nodes[2].TriggerElection(ctx)
		h.net.Nodes[3].TriggerElection(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[0])
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, h.net.Nodes[1])

		// selecting node 2 as the leader
		h.net.Nodes[0].TriggerElection(ctx)
		h.net.Nodes[1].TriggerElection(ctx)
		h.net.Nodes[3].TriggerElection(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[1])
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, h.net.Nodes[2])

		// selecting node 3 as the leader
		h.net.Nodes[0].TriggerElection(ctx)
		h.net.Nodes[1].TriggerElection(ctx)
		h.net.Nodes[2].TriggerElection(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[2])
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, h.net.Nodes[3])

		// back to node 0 as the leader
		h.net.Nodes[1].TriggerElection(ctx)
		h.net.Nodes[2].TriggerElection(ctx)
		h.net.Nodes[3].TriggerElection(ctx)

		h.net.ResumeRequestNewBlockOnNodes(ctx, h.net.Nodes[3])
		h.net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, h.net.Nodes[0])
	})
}
