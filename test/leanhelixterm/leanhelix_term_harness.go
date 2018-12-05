package leanhelixterm

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

type harness struct {
	t                 *testing.T
	myPublicKey       primitives.Ed25519PublicKey
	keyManager        *builders.MockKeyManager
	myNode            *builders.Node
	net               *builders.TestNetwork
	term              *leanhelix.LeanHelixTerm
	storage           leanhelix.Storage
	electionTrigger   *builders.ElectionTriggerMock
	failVerifications bool
}

func NewHarness(ctx context.Context, t *testing.T, blocksPool ...leanhelix.Block) *harness {
	net := builders.NewTestNetworkBuilder().WithNodeCount(4).WithBlocksPool(blocksPool).Build()
	myNode := net.Nodes[0]
	keyManager := myNode.KeyManager
	termConfig := myNode.BuildConfig()
	term := leanhelix.NewLeanHelixTerm(ctx, termConfig, nil, myNode.GetLatestBlock())
	term.StartTerm(ctx)

	return &harness{
		t:                 t,
		myPublicKey:       keyManager.MyPublicKey(),
		myNode:            myNode,
		net:               net,
		keyManager:        myNode.KeyManager,
		term:              term,
		storage:           termConfig.Storage,
		electionTrigger:   myNode.ElectionTrigger,
		failVerifications: false,
	}
}

func (h *harness) failValidations() {
	h.myNode.BlockUtils.ValidationResult = false
}

func (h *harness) checkView(expectedView primitives.View) {
	view := h.term.GetView()
	require.Equal(h.t, expectedView, view, fmt.Sprintf("Term should have view=%d, but got %d", expectedView, view))
}

func (h *harness) triggerElection(ctx context.Context) {
	h.electionTrigger.ManualTriggerSync(ctx)
}

func (h *harness) getMyNodePk() primitives.Ed25519PublicKey {
	return h.getMemberPk(0)
}

func (h *harness) getMemberPk(nodeIdx int) primitives.Ed25519PublicKey {
	return h.net.Nodes[nodeIdx].KeyManager.MyPublicKey()
}

func (h *harness) electionTillView(ctx context.Context, view primitives.View) {
	for {
		if h.term.GetView() == view {
			break
		}
		h.triggerElection(ctx)
	}
}

func (h *harness) setNode1AsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	h.receiveNewView(ctx, 1, blockHeight, view, block)
}

func (h *harness) setMeAsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	h.receiveNewView(ctx, 0, blockHeight, view, block)
}

func (h *harness) receiveViewChange(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[fromNodeIdx]
	vc := builders.AViewChangeMessage(sender.KeyManager, blockHeight, view, nil)
	h.term.HandleLeanHelixViewChange(ctx, vc)
}

func (h *harness) receivePreprepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	leader := h.net.Nodes[fromNode]
	ppm := builders.APreprepareMessage(leader.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrePrepare(ctx, ppm)
}

func (h *harness) receivePreprepareMessage(ctx context.Context, ppm *leanhelix.PreprepareMessage) {
	h.term.HandleLeanHelixPrePrepare(ctx, ppm)
}

func (h *harness) receivePrepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[fromNode]
	pm := builders.APrepareMessage(sender.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrepare(ctx, pm)
}

func (h *harness) createPreprepareMessage(fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block, blockHash primitives.Uint256) *leanhelix.PreprepareMessage {
	leader := h.net.Nodes[fromNode]
	messageFactory := leanhelix.NewMessageFactory(leader.KeyManager)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, blockHash)
}

func (h *harness) receiveCustomNewViewMessage(
	ctx context.Context,
	leaderNode int,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block,
	preprepareBlock leanhelix.Block,
	preprepareBlockHeight primitives.BlockHeight,
	preprepareView primitives.View,
	vcsBlockHeight [3]primitives.BlockHeight,
	vcsView [3]primitives.View) {

	var members []*builders.Node
	for i, node := range h.net.Nodes {
		if i != leaderNode {
			members = append(members, node)
		}
	}

	newLeader := h.net.Nodes[leaderNode]
	ppmFactory := leanhelix.NewMessageFactory(newLeader.KeyManager)
	ppmCB := ppmFactory.CreatePreprepareMessageContentBuilder(preprepareBlockHeight, preprepareView, preprepareBlock, builders.CalculateBlockHash(preprepareBlock))

	var votes []*leanhelix.ViewChangeMessageContentBuilder
	for idx, voter := range members {
		messageFactory := leanhelix.NewMessageFactory(voter.KeyManager)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(vcsBlockHeight[idx], vcsView[idx], nil)
		votes = append(votes, vcmCB)
	}

	messageFactory := leanhelix.NewMessageFactory(newLeader.KeyManager)
	nvcb := messageFactory.CreateNewViewMessageContentBuilder(blockHeight, view, ppmCB, votes)
	nvm := leanhelix.NewNewViewMessage(nvcb.Build(), block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) receiveNewViewMessageWithDuplicateVotes(
	ctx context.Context,
	leaderNode int,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) {

	var members []*builders.Node
	for i, node := range h.net.Nodes {
		if i != leaderNode {
			members = append(members, node)
		}
	}

	newLeader := h.net.Nodes[leaderNode]
	ppmFactory := leanhelix.NewMessageFactory(newLeader.KeyManager)
	ppmCB := ppmFactory.CreatePreprepareMessageContentBuilder(blockHeight, view, block, builders.CalculateBlockHash(block))

	var votes []*leanhelix.ViewChangeMessageContentBuilder
	for _, voter := range members {
		messageFactory := leanhelix.NewMessageFactory(voter.KeyManager)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(blockHeight, view, nil)
		votes = append(votes, vcmCB)
	}

	// override vote 1 with vote 2, so vote 1 will be twice
	votes[1] = votes[2]

	messageFactory := leanhelix.NewMessageFactory(newLeader.KeyManager)
	nvcb := messageFactory.CreateNewViewMessageContentBuilder(blockHeight, view, ppmCB, votes)
	nvm := leanhelix.NewNewViewMessage(nvcb.Build(), block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) receiveNewView(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	nvcb := h.createNewViewContentBuilder(fromNodeIdx, blockHeight, view, block)
	nvm := leanhelix.NewNewViewMessage(nvcb.Build(), block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) createNewViewContentBuilder(
	fromNode int,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.NewViewMessageContentBuilder {

	var members []*builders.Node
	for i, node := range h.net.Nodes {
		if i != fromNode {
			members = append(members, node)
		}
	}

	leader := h.net.Nodes[fromNode]
	return builders.ANewViewContentBuilder(leader, members, blockHeight, view, block)
}

func (h *harness) getLastSentViewChangeMessage() *leanhelix.ViewChangeMessage {
	messages := h.myNode.Gossip.GetSentMessages(leanhelix.LEAN_HELIX_VIEW_CHANGE)
	lastMessage := messages[len(messages)-1].ToConsensusMessage()
	return lastMessage.(*leanhelix.ViewChangeMessage)
}

func (h *harness) countViewChange(blockHeight primitives.BlockHeight, view primitives.View) int {
	messages, _ := h.storage.GetViewChangeMessages(blockHeight, view)
	return len(messages)
}

func (h *harness) countCommits(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) int {
	messages, _ := h.storage.GetCommitMessages(blockHeight, view, builders.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) hasPreprepare(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) bool {
	message, ok := h.storage.GetPreprepareMessage(blockHeight, view)

	if message == nil || !ok {
		return false
	}

	return builders.BlocksAreEqual(message.Block(), block)
}

func (h *harness) countPrepare(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) int {
	messages, _ := h.storage.GetPrepareMessages(blockHeight, view, builders.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) failFutureVerifications() {
	h.keyManager.FailFutureVerifications = true
}

func (h *harness) disposeTerm() {
	h.term.Dispose()
}
