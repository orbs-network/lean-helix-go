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
	term            *leanhelix.LeanHelixTerm
	filter          *leanhelix.ConsensusMessageFilter
	electionTrigger *builders.ElectionTriggerMock
	blockUtils      *builders.MockBlockUtils
}

func NewHarness(t *testing.T) *harness {
	net := builders.ABasicTestNetwork()
	node := net.Nodes[0]
	termConfig := node.BuildConfig()
	node.ElectionTrigger.PauseOnTick = true

	// term initialization
	filter := leanhelix.NewConsensusMessageFilter(termConfig.KeyManager.MyPublicKey())
	term := leanhelix.NewLeanHelixTerm(termConfig, filter, 0)

	return &harness{
		t:               t,
		term:            term,
		filter:          filter,
		electionTrigger: node.ElectionTrigger,
		blockUtils:      node.BlockUtils,
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

func (h *harness) sendLeaderChange(ctx context.Context, view primitives.View) {
	block := builders.CreateBlock(builders.GenesisBlock)
	publicKey := primitives.Ed25519PublicKey("Dummy PublicKey")
	keyManager := builders.NewMockKeyManager(publicKey)
	nvm := builders.ANewViewMessage(keyManager, 0, view, nil, nil, block)
	go h.filter.OnGossipMessage(ctx, nvm.ToConsensusRawMessage())
}

func (h *harness) verifyViewDoesNotChange(view primitives.View) {
}
