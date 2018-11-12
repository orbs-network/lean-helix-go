package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/require"
	"testing"
)

type harness struct {
	t               *testing.T
	term            *leanhelix.LeanHelixTerm
	electionTrigger *builders.ElectionTriggerMock
}

func NewHarness(t *testing.T) *harness {
	publicKey := primitives.Ed25519PublicKey("My PublicKey")
	keyManager := builders.NewMockKeyManager(publicKey)
	discovery := gossip.NewGossipDiscovery()
	gossip := gossip.NewGossip(discovery)
	discovery.RegisterGossip(publicKey, gossip)
	blockUtils := builders.NewMockBlockUtils(nil)
	electionTrigger := builders.NewMockElectionTrigger()
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
		electionTrigger: electionTrigger,
	}
}

func (h *harness) startConsensus(ctx context.Context) {
	go h.term.WaitForBlock(ctx)
}

func (h *harness) verifyView(view primitives.View) {
	require.Equal(h.t, primitives.View(0), h.term.GetView(), "Term should have view=0 on init")
}

func (h *harness) waitForView(view primitives.View) {
	test.Eventually(1000, func() bool { return primitives.View(1) == h.term.GetView() })
}

func (h *harness) changeLeader() {
	h.electionTrigger.Trigger()
}
