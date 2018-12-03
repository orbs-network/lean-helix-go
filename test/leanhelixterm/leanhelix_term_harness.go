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

func (h *harness) setNode1AsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	me := h.net.Nodes[0]
	leader := h.net.Nodes[1]
	node2 := h.net.Nodes[2]
	node3 := h.net.Nodes[3]
	members := []*builders.Node{me, node2, node3}
	nvm := builders.AValidNewViewMessage(leader, members, blockHeight, view, block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) setMeAsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	h.sendNewView(ctx, 0, blockHeight, view, block)
}

func (h *harness) sendViewChange(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[fromNodeIdx]
	vc := builders.AViewChangeMessage(sender.KeyManager, blockHeight, view, nil)
	h.term.HandleLeanHelixViewChange(ctx, vc)
}

func (h *harness) sendPreprepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	leader := h.net.Nodes[fromNode]
	ppm := builders.APreprepareMessage(leader.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrePrepare(ctx, ppm)
}

func (h *harness) sendPreprepareWithSpecificBlockHash(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block, blockHash primitives.Uint256) {
	leader := h.net.Nodes[fromNode]
	messageFactory := leanhelix.NewMessageFactory(leader.KeyManager)
	ppm := messageFactory.CreatePreprepareMessage(blockHeight, view, block, blockHash)
	h.term.HandleLeanHelixPrePrepare(ctx, ppm)
}

func (h *harness) sendPrepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[fromNode]
	pm := builders.APrepareMessage(sender.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrepare(ctx, pm)
}

func (h *harness) sendNewView(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	var members []*builders.Node
	for i, node := range h.net.Nodes {
		if i != fromNodeIdx {
			members = append(members, node)
		}
	}

	leaderNode := h.net.Nodes[fromNodeIdx]
	nvm := builders.AValidNewViewMessage(leaderNode, members, blockHeight, view, block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
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
