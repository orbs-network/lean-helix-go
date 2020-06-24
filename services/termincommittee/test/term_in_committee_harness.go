// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/blockreferencetime"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

const TERM_IN_COMMITTEE_HARNESS_LOGS_TO_CONSOLE = false

type harness struct {
	t               *testing.T
	instanceId      primitives.InstanceId
	myMemberId      primitives.MemberId
	keyManager      *mocks.MockKeyManager
	myNode          *network.Node
	net             *network.TestNetwork
	termInCommittee *termincommittee.TermInCommittee
	storage         interfaces.Storage
	electionTrigger interfaces.ElectionScheduler
}

func NewHarness(ctx context.Context, t *testing.T, blocksPool ...interfaces.Block) *harness {
	return NewHarnessForNodeInd(ctx, 0, t, blocksPool)
}

func NewHarnessForNodeInd(ctx context.Context, nodeInd int, t *testing.T, blocksPool []interfaces.Block) *harness {
	net := network.
		NewTestNetworkBuilder().
		WithNodeCount(4).
		WithNodeWeights([]primitives.MemberWeight{1, 2, 3, 4}).
		WithBlocks(blocksPool...).
		//LogToConsole().
		Build(ctx)
	myNode := net.Nodes[nodeInd]
	var logOutput interfaces.Logger
	if TERM_IN_COMMITTEE_HARNESS_LOGS_TO_CONSOLE {
		logOutput = logger.NewConsoleLogger(test.NameHashPrefix(t, 4))
	} else {
		logOutput = logger.NewSilentLogger()
	}
	termConfig := myNode.BuildConfig(logOutput)
	log := logger.NewLhLogger(termConfig, mocks.NewMockState().State)

	prevBlock := myNode.GetLatestBlock()
	state := mocks.NewMockState().WithHeightView(blockheight.GetBlockHeight(prevBlock)+1, 0)
	committeeMembers, _ := termConfig.Membership.RequestOrderedCommittee(ctx, state.Height(), uint64(12345), blockreferencetime.GetBlockReferenceTime(prevBlock))
	messageFactory := messagesfactory.NewMessageFactory(termConfig.InstanceId, termConfig.KeyManager, termConfig.Membership.MyMemberId(), 0)
	log.Info("NewHarness calling NewTermInCommittee with H=%d", state.Height())

	// TODO state.State is shadowing state.State and is generally meaninless
	termInCommittee := termincommittee.NewTermInCommittee(log, termConfig, state.State, messageFactory, myNode.ElectionTrigger, committeeMembers, prevBlock, true, nil)

	return &harness{
		t:               t,
		instanceId:      termConfig.InstanceId,
		myMemberId:      myNode.MemberId,
		myNode:          myNode,
		net:             net,
		keyManager:      myNode.KeyManager,
		termInCommittee: termInCommittee,
		storage:         termConfig.Storage,
		electionTrigger: myNode.ElectionTrigger,
	}
}

func (h *harness) failMyNodeBlockProposalValidations() {
	h.myNode.BlockUtils.(*mocks.PausableBlockUtils).WithFailingBlockProposalValidations()
}

func (h *harness) assertView(expectedView primitives.View) {
	view := h.termInCommittee.State.View()
	require.Equal(h.t, expectedView, view, fmt.Sprintf("TermInCommittee should have view=%d, but got %d", uint64(expectedView), uint64(view)))
}

func (h *harness) triggerElection(ctx context.Context) {
	electionTriggerMock, ok := h.electionTrigger.(*mocks.ElectionTriggerMock)
	if ok {
		electionTriggerMock.ManualTrigger(ctx, h.myNode.State().HeightView())
	} else {
		panic("You are trying to trigger election with an election trigger that is not the ElectionTriggerMock")
	}

	electionTriggerMock.InvokeElectionHandler()
}

func (h *harness) getMyNodeMemberId() primitives.MemberId {
	return h.getNodeMemberId(0)
}

func (h *harness) getNodeMemberId(nodeIdx int) primitives.MemberId {
	return h.net.Nodes[nodeIdx].MemberId
}

func (h *harness) getMyKeyManager() interfaces.KeyManager {
	return h.getMemberKeyManager(0)
}

func (h *harness) getMemberKeyManager(nodeIdx int) interfaces.KeyManager {
	return h.net.Nodes[nodeIdx].KeyManager
}

func (h *harness) builderMessageSender(nodeIdx int) *builders.MessageSigner {
	node := h.net.Nodes[nodeIdx]
	return &builders.MessageSigner{KeyManager: node.KeyManager, MemberId: node.MemberId}
}

func (h *harness) builderMessageSenders(nodesIds ...int) []*builders.MessageSigner {
	result := make([]*builders.MessageSigner, len(nodesIds))
	for _, idx := range nodesIds {
		result = append(result, h.builderMessageSender(idx))
	}
	return result
}

func (h *harness) electionTillView(ctx context.Context, view primitives.View) {
	for {
		if h.termInCommittee.State.View() == view {
			break
		}
		h.triggerElection(ctx)
	}
}

func (h *harness) setNode1AsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	h.receiveAndHandleNewView(ctx, 1, blockHeight, view, block)
}

func (h *harness) setMeAsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	h.receiveAndHandleNewView(ctx, 0, blockHeight, view, block)
}

func (h *harness) receiveAndHandlePreprepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	leader := h.net.Nodes[fromNode]
	ppm := builders.APreprepareMessage(h.instanceId, leader.KeyManager, leader.MemberId, blockHeight, view, block)
	h.termInCommittee.HandlePrePrepare(ppm)
}

func (h *harness) receiveAndHandlePrepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	sender := h.net.Nodes[fromNode]
	pm := builders.APrepareMessage(h.instanceId, sender.KeyManager, sender.MemberId, blockHeight, view, block)
	h.termInCommittee.HandlePrepare(pm)
}

func (h *harness) receiveAndHandleViewChange(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View) {
	sender := h.net.Nodes[fromNodeIdx]
	vc := builders.AViewChangeMessage(h.instanceId, sender.KeyManager, sender.MemberId, blockHeight, view, nil)
	h.termInCommittee.HandleViewChange(vc)
}

func (h *harness) receiveAndHandleNewView(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	leaderKeyManager := h.getMemberKeyManager(fromNodeIdx)
	leaderMemberId := h.getNodeMemberId(fromNodeIdx)

	var voters []*builders.Voter
	for i, node := range h.net.Nodes {
		if i != fromNodeIdx {
			voters = append(voters, &builders.Voter{KeyManager: node.KeyManager, MemberId: node.MemberId})
		}
	}

	votes := builders.ASimpleViewChangeVotes(h.instanceId, voters, blockHeight, view)
	nvm := builders.
		NewNewViewBuilder().
		LeadBy(leaderKeyManager, leaderMemberId).
		WithViewChangeVotes(votes).
		OnBlock(block).
		OnBlockHeight(blockHeight).
		OnView(view).
		Build()
	h.termInCommittee.HandleNewView(nvm)
}

func (h *harness) handleViewChangeMessage(ctx context.Context, msg *interfaces.ViewChangeMessage) {
	h.termInCommittee.HandleViewChange(msg)
}

func (h *harness) handleNewViewMessage(ctx context.Context, nvm *interfaces.NewViewMessage) {
	h.termInCommittee.HandleNewView(nvm)
}

func (h *harness) createPreprepareMessage(fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block, blockHash primitives.BlockHash) *interfaces.PreprepareMessage {
	leader := h.net.Nodes[fromNode]
	messageFactory := messagesfactory.NewMessageFactory(h.instanceId, leader.KeyManager, leader.MemberId, 0)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, blockHash)
}

func (h *harness) getLastSentViewChangeMessage() *interfaces.ViewChangeMessage {
	messages := h.myNode.Communication.GetSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE)
	lastMessage := interfaces.ToConsensusMessage(messages[len(messages)-1])
	return lastMessage.(*interfaces.ViewChangeMessage)
}

func (h *harness) countPrepare(blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) int {
	messages, _ := h.storage.GetPrepareMessages(blockHeight, view, mocks.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) countCommits(blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) int {
	messages, _ := h.storage.GetCommitMessages(blockHeight, view, mocks.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) countViewChange(blockHeight primitives.BlockHeight, view primitives.View) int {
	messages, _ := h.storage.GetViewChangeMessages(blockHeight, view)
	return len(messages)
}

func (h *harness) hasPreprepare(blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) bool {
	message, ok := h.storage.GetPreprepareMessage(blockHeight, view)

	if message == nil || !ok {
		return false
	}

	return matchers.BlocksAreEqual(message.Block(), block)
}

func (h *harness) failFutureVerifications() {
	h.keyManager.FailFutureVerifications = true
}

func (h *harness) disposeTerm() {
	h.termInCommittee.Dispose()
}
