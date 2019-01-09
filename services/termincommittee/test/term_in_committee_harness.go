package test

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

type harness struct {
	t                 *testing.T
	instanceId        primitives.InstanceId
	myMemberId        primitives.MemberId
	keyManager        *mocks.MockKeyManager
	myNode            *network.Node
	net               *network.TestNetwork
	termInCommittee   *termincommittee.TermInCommittee
	storage           interfaces.Storage
	electionTrigger   *mocks.ElectionTriggerMock
	failVerifications bool
}

func NewHarness(ctx context.Context, t *testing.T, blocksPool ...interfaces.Block) *harness {
	net := network.NewTestNetworkBuilder().WithNodeCount(4).WithBlocks(blocksPool).Build()
	myNode := net.Nodes[0]
	termConfig := myNode.BuildConfig(nil)

	prevBlock := myNode.GetLatestBlock()
	blockHeight := blockheight.GetBlockHeight(prevBlock) + 1
	committeeMembers := termConfig.Membership.RequestOrderedCommittee(ctx, blockHeight, uint64(12345))
	messageFactory := messagesfactory.NewMessageFactory(termConfig.InstanceId, termConfig.KeyManager, termConfig.Membership.MyMemberId(), 0)
	termInCommittee := termincommittee.NewTermInCommittee(ctx, termConfig, messageFactory, committeeMembers, blockHeight, prevBlock, nil)

	return &harness{
		t:                 t,
		instanceId:        termConfig.InstanceId,
		myMemberId:        myNode.MemberId,
		myNode:            myNode,
		net:               net,
		keyManager:        myNode.KeyManager,
		termInCommittee:   termInCommittee,
		storage:           termConfig.Storage,
		electionTrigger:   myNode.ElectionTrigger,
		failVerifications: false,
	}
}

func (h *harness) failValidations() {
	h.myNode.BlockUtils.ValidationResult = false
}

func (h *harness) checkView(expectedView primitives.View) {
	view := h.termInCommittee.GetView()
	require.Equal(h.t, expectedView, view, fmt.Sprintf("TermInCommittee should have view=%d, but got %d", expectedView, view))
}

func (h *harness) triggerElection(ctx context.Context) {
	h.electionTrigger.ManualTriggerSync(ctx)
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
		if h.termInCommittee.GetView() == view {
			break
		}
		h.triggerElection(ctx)
	}
}

func (h *harness) setNode1AsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	h.receiveNewView(ctx, 1, blockHeight, view, block)
}

func (h *harness) setMeAsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	h.receiveNewView(ctx, 0, blockHeight, view, block)
}

func (h *harness) receiveViewChange(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	sender := h.net.Nodes[fromNodeIdx]
	vc := builders.AViewChangeMessage(h.instanceId, sender.KeyManager, sender.MemberId, blockHeight, view, nil)
	h.termInCommittee.HandleViewChange(ctx, vc)
}

func (h *harness) receiveViewChangeMessage(ctx context.Context, msg *interfaces.ViewChangeMessage) {
	h.termInCommittee.HandleViewChange(ctx, msg)
}

func (h *harness) receivePreprepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	leader := h.net.Nodes[fromNode]
	ppm := builders.APreprepareMessage(h.instanceId, leader.KeyManager, leader.MemberId, blockHeight, view, block)
	h.termInCommittee.HandlePrePrepare(ctx, ppm)
}

func (h *harness) receivePreprepareMessage(ctx context.Context, ppm *interfaces.PreprepareMessage) {
	h.termInCommittee.HandlePrePrepare(ctx, ppm)
}

func (h *harness) receivePrepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	sender := h.net.Nodes[fromNode]
	pm := builders.APrepareMessage(h.instanceId, sender.KeyManager, sender.MemberId, blockHeight, view, block)
	h.termInCommittee.HandlePrepare(ctx, pm)
}

func (h *harness) createPreprepareMessage(fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block, blockHash primitives.BlockHash) *interfaces.PreprepareMessage {
	leader := h.net.Nodes[fromNode]
	messageFactory := messagesfactory.NewMessageFactory(h.instanceId, leader.KeyManager, leader.MemberId, 0)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, blockHash)
}

func (h *harness) HandleNewView(ctx context.Context, nvm *interfaces.NewViewMessage) {
	h.termInCommittee.HandleNewView(ctx, nvm)
}

func (h *harness) receiveNewView(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) {
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
	h.termInCommittee.HandleNewView(ctx, nvm)
}

func (h *harness) getLastSentViewChangeMessage() *interfaces.ViewChangeMessage {
	messages := h.myNode.Communication.GetSentMessages(protocol.LEAN_HELIX_VIEW_CHANGE)
	lastMessage := interfaces.ToConsensusMessage(messages[len(messages)-1])
	return lastMessage.(*interfaces.ViewChangeMessage)
}

func (h *harness) countViewChange(blockHeight primitives.BlockHeight, view primitives.View) int {
	messages, _ := h.storage.GetViewChangeMessages(blockHeight, view)
	return len(messages)
}

func (h *harness) countCommits(blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) int {
	messages, _ := h.storage.GetCommitMessages(blockHeight, view, mocks.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) hasPreprepare(blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) bool {
	message, ok := h.storage.GetPreprepareMessage(blockHeight, view)

	if message == nil || !ok {
		return false
	}

	return matchers.BlocksAreEqual(message.Block(), block)
}

func (h *harness) countPrepare(blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) int {
	messages, _ := h.storage.GetPrepareMessages(blockHeight, view, mocks.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) failFutureVerifications() {
	h.keyManager.FailFutureVerifications = true
}

func (h *harness) disposeTerm() {
	h.termInCommittee.Dispose()
}
