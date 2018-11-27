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
	filter          *leanhelix.ConsensusMessageFilter
	storage         leanhelix.Storage
	electionTrigger *builders.ElectionTriggerMock
}

func NewHarness(t *testing.T) *harness {
	net := builders.ABasicTestNetwork()
	node := net.Nodes[0]
	myPublicKey := node.KeyManager.MyPublicKey()
	termConfig := node.BuildConfig()
	filter := leanhelix.NewConsensusMessageFilter(myPublicKey)
	term := leanhelix.NewLeanHelixTerm(termConfig, nil, node.GetLatestBlock().Height()+1)

	return &harness{
		t:               t,
		myPublicKey:     myPublicKey,
		net:             net,
		term:            term,
		filter:          filter,
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

func (h *harness) sendNewView(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	leader := h.net.Nodes[0]
	node1 := h.net.Nodes[1]
	node2 := h.net.Nodes[2]
	node3 := h.net.Nodes[3]
	members := []*builders.Node{node1, node2, node3}
	nvm := builders.AValidNewViewMessage(leader, members, blockHeight, view, block)
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) sendViewChange(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[3]
	vc := builders.AViewChangeMessage(sender.KeyManager, blockHeight, view, nil)
	h.term.HandleLeanHelixViewChange(ctx, vc)
}

func (h *harness) countViewChange(blockHeight primitives.BlockHeight, view primitives.View) int {
	return len(h.storage.GetViewChangeMessages(blockHeight, view))
}
