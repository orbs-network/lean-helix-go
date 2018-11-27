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
	net             *builders.TestNetwork
	term            *leanhelix.LeanHelixTerm
	filter          *leanhelix.ConsensusMessageFilter
	electionTrigger *builders.ElectionTriggerMock
	keyManager      leanhelix.KeyManager
	blockUtils      *builders.MockBlockUtils
	storage         leanhelix.Storage
}

func NewHarness(t *testing.T) *harness {
	net := builders.ABasicTestNetwork()
	node := net.Nodes[1]
	termConfig := node.BuildConfig()
	lastCommittedBlock := node.GetLatestBlock()

	// term initialization
	filter := leanhelix.NewConsensusMessageFilter(termConfig.KeyManager.MyPublicKey())
	term := leanhelix.NewLeanHelixTerm(termConfig, nil, lastCommittedBlock.Height()+1)

	h := &harness{
		t:               t,
		net:             net,
		term:            term,
		filter:          filter,
		electionTrigger: node.ElectionTrigger,
		blockUtils:      node.BlockUtils,
		keyManager:      node.KeyManager,
		storage:         node.Storage,
	}
	return h
}

func (h *harness) startConsensus(ctx context.Context) {
}

func (h *harness) waitForView(expectedView primitives.View) {
	view := h.term.GetView()
	require.Equal(h.t, view, expectedView, fmt.Sprintf("Term should have view=%d, but got %d", expectedView, view))
}

func (h *harness) triggerElection() {
	h.electionTrigger.ManualTrigger()
}

func (h *harness) sendLeaderChanged(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	leader := h.net.Nodes[0]
	node1 := h.net.Nodes[1]
	node2 := h.net.Nodes[2]
	node3 := h.net.Nodes[3]
	members := []*builders.Node{node1, node2, node3}
	nvm := builders.AValidNewViewMessage(leader, members, blockHeight, view, block)
	go h.filter.GossipMessageReceived(ctx, nvm.ToConsensusRawMessage())
}

func (h *harness) sendChangeLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[3]
	vc := builders.AViewChangeMessage(sender.KeyManager, blockHeight, view, nil)
	go h.filter.GossipMessageReceived(ctx, vc.ToConsensusRawMessage())
}

func (h *harness) countViewChange(blockHeight primitives.BlockHeight, view primitives.View) int {
	return len(h.storage.GetViewChangeMessages(blockHeight, view))
}
