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
}

func NewHarness(t *testing.T) *harness {
	net := builders.ABasicTestNetwork()
	node := net.Nodes[1]
	termConfig := node.BuildConfig()
	node.ElectionTrigger.PauseOnTick = true
	lastCommittedBlock := node.GetLatestBlock()

	// term initialization
	filter := leanhelix.NewConsensusMessageFilter(termConfig.KeyManager.MyPublicKey())
	term := leanhelix.NewLeanHelixTerm(termConfig, filter, lastCommittedBlock.Height()+1)

	return &harness{
		t:               t,
		net:             net,
		term:            term,
		filter:          filter,
		electionTrigger: node.ElectionTrigger,
		blockUtils:      node.BlockUtils,
		keyManager:      node.KeyManager,
	}
}

func (h *harness) startConsensus(ctx context.Context) {
	go h.term.WaitForBlock(ctx)
}

func (h *harness) waitForView(expectedView primitives.View) {
	h.electionTrigger.TickSns.WaitForSignal()
	view := h.term.GetView()
	h.electionTrigger.TickSns.Resume()
	require.Equal(h.t, view, expectedView, fmt.Sprintf("Term should have view=%d, but got %d", expectedView, view))
}

func (h *harness) triggerElection() {
	h.electionTrigger.ManualTrigger()
}

func (h *harness) sendLeaderChange(ctx context.Context, view primitives.View, block leanhelix.Block) {
	leader := h.net.Nodes[0]
	node1 := h.net.Nodes[1]
	node2 := h.net.Nodes[2]
	node3 := h.net.Nodes[3]
	members := []*builders.Node{node1, node2, node3}
	nvm := builders.AValidNewViewMessage(leader, members, 1, view, block)
	go h.filter.OnGossipMessage(ctx, nvm.ToConsensusRawMessage())
}
