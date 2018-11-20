package leanhelixterm

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/require"
	"testing"
)

type harness struct {
	t               *testing.T
	term            *leanhelix.LeanHelixTerm
	filter          *leanhelix.ConsensusMessageFilter
	electionTrigger *builders.ElectionTriggerMock
}

func NewHarness(t *testing.T) *harness {
	publicKey := primitives.Ed25519PublicKey("My PublicKey")
	keyManager := builders.NewMockKeyManager(publicKey)
	discovery := gossip.NewGossipDiscovery()
	gossip := gossip.NewGossip(discovery)
	discovery.RegisterGossip(publicKey, gossip)
	blockUtils := builders.NewMockBlockUtils(nil)
	electionTrigger := builders.NewMockElectionTrigger(true)
	storage := leanhelix.NewInMemoryStorage()
	termConfig := &leanhelix.Config{
		NetworkCommunication: gossip,
		BlockUtils:           blockUtils,
		KeyManager:           keyManager,
		ElectionTrigger:      electionTrigger,
		Storage:              storage,
	}
	filter := leanhelix.NewConsensusMessageFilter(publicKey)
	term := leanhelix.NewLeanHelixTerm(termConfig, filter, 0)

	return &harness{
		t:               t,
		term:            term,
		filter:          filter,
		electionTrigger: electionTrigger,
	}
}

func (h *harness) startConsensus(ctx context.Context) {
	go h.term.WaitForBlock(ctx)
}

func (h *harness) waitForView(expectedView primitives.View) {
	view := h.electionTrigger.WaitForNextView()
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
