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
	t               *testing.T
	myPublicKey     primitives.Ed25519PublicKey
	net             *builders.TestNetwork
	term            *leanhelix.LeanHelixTerm
	storage         leanhelix.Storage
	electionTrigger *builders.ElectionTriggerMock
}

func NewHarness(ctx context.Context, t *testing.T) *harness {
	net := builders.ABasicTestNetwork()
	node := net.Nodes[0]
	myPublicKey := node.KeyManager.MyPublicKey()
	termConfig := node.BuildConfig()
	term := leanhelix.NewLeanHelixTerm(ctx, termConfig, nil, node.GetLatestBlock().Height()+1)

	return &harness{
		t:               t,
		myPublicKey:     myPublicKey,
		net:             net,
		term:            term,
		storage:         termConfig.Storage,
		electionTrigger: node.ElectionTrigger,
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
	me := h.net.Nodes[0]
	node1 := h.net.Nodes[1]
	node2 := h.net.Nodes[2]
	node3 := h.net.Nodes[3]
	members := []*builders.Node{node1, node2, node3}
	nvm := builders.AValidNewViewMessage(me, members, blockHeight, view, block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) sendViewChange(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[3]
	vc := builders.AViewChangeMessage(sender.KeyManager, blockHeight, view, nil)
	h.term.HandleLeanHelixViewChange(ctx, vc)
}

func (h *harness) sendPreprepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	leader := h.net.Nodes[fromNode]
	ppm := builders.APreprepareMessage(leader.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrePrepare(ctx, ppm)
}

func (h *harness) sendPrepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[fromNode]
	pm := builders.APrepareMessage(sender.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrepare(ctx, pm)
}

func (h *harness) countViewChange(blockHeight primitives.BlockHeight, view primitives.View) int {
	messages, _ := h.storage.GetViewChangeMessages(blockHeight, view)
	return len(messages)
}

func (h *harness) countCommits(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) int {
	messages, _ := h.storage.GetCommitMessages(blockHeight, view, builders.CalculateBlockHash(block))
	return len(messages)
}
