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
	net               *builders.TestNetwork
	term              *leanhelix.LeanHelixTerm
	storage           leanhelix.Storage
	electionTrigger   *builders.ElectionTriggerMock
	failVerifications bool
}

func NewHarness(ctx context.Context, t *testing.T) *harness {
	net := builders.ABasicTestNetwork()
	myNode := net.Nodes[0]
	keyManager := myNode.KeyManager
	termConfig := myNode.BuildConfig()
	term := leanhelix.NewLeanHelixTerm(ctx, termConfig, nil, myNode.GetLatestBlock())

	return &harness{
		t:                 t,
		myPublicKey:       keyManager.MyPublicKey(),
		net:               net,
		keyManager:        myNode.KeyManager,
		term:              term,
		storage:           termConfig.Storage,
		electionTrigger:   myNode.ElectionTrigger,
		failVerifications: false,
	}
}

func (h *harness) checkView(expectedView primitives.View) {
	view := h.term.GetView()
	require.Equal(h.t, expectedView, view, fmt.Sprintf("Term should have view=%d, but got %d", expectedView, view))
}

func (h *harness) triggerElection(ctx context.Context) {
	h.electionTrigger.ManualTriggerSync(ctx)
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

func (h *harness) receiveNewView(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	nvcb := h.createNewViewContentBuilder(fromNodeIdx, blockHeight, view, block, builders.CalculateBlockHash(block))
	nvm := leanhelix.NewNewViewMessage(nvcb.Build(), block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) createPreprepareMessage(fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block, blockHash primitives.Uint256) *leanhelix.PreprepareMessage {
	leader := h.net.Nodes[fromNode]
	messageFactory := leanhelix.NewMessageFactory(leader.KeyManager)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, blockHash)
}

func (h *harness) createNewViewContentBuilder(
	fromNode int,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block,
	blockHash primitives.Uint256) *leanhelix.NewViewMessageContentBuilder {

	var members []*builders.Node
	for i, node := range h.net.Nodes {
		if i != fromNode {
			members = append(members, node)
		}
	}

	leader := h.net.Nodes[fromNode]
	return builders.ANewViewContentBuilder(leader, members, blockHeight, view, block)
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
