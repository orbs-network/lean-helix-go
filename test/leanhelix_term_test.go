package test

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestViewIncrementedAfterElectionTrigger(t *testing.T) {
	WithContext(func(ctx context.Context) {
		publicKey := primitives.Ed25519PublicKey("My PublicKey")
		keyManager := builders.NewMockKeyManager(publicKey)
		discovery := gossip.NewGossipDiscovery()
		gossip := gossip.NewGossip(discovery)
		discovery.RegisterGossip(publicKey, gossip)
		blockUtils := builders.NewMockBlockUtils(nil)
		electionTrigger := builders.NewMockElectionTrigger()
		storage := lh.NewInMemoryStorage()
		termConfig := &lh.Config{
			NetworkCommunication: gossip,
			BlockUtils:           blockUtils,
			KeyManager:           keyManager,
			ElectionTrigger:      electionTrigger,
			Storage:              storage,
		}
		filter := lh.NewConsensusMessageFilter(publicKey)
		term := lh.NewLeanHelixTerm(termConfig, filter, 0)
		go term.WaitForBlock(ctx)

		require.Equal(t, primitives.View(0), term.GetView(), "Term should have view=0 on init")
		electionTrigger.Trigger()
		Eventually(1000, func() bool { return primitives.View(1) == term.GetView() })
	})
}

// TODO: uncomment
//func TestRejectNewViewMessagesFromPast(t *testing.T) {
//	WithContext(func(ctx context.Context) {
//		height := BlockHeight(0)
//		view := View(0)
//		block := builders.CreateBlock(builders.GenesisBlock)
//		net := builders.ABasicTestNetwork()
//
//		node := net.Nodes[0]
//		messageFactory := lh.NewMessageFactory(node.KeyManager)
//		ppmContentBuilder := messageFactory.CreatePreprepareMessageContentBuilder(height, view, block)
//		termConfig := net.Nodes[0].BuildConfig()
//		filter := lh.NewConsensusMessageFilter(termConfig.KeyManager.MyPublicKey())
//		term := lh.NewLeanHelixTerm(termConfig, filter, 0)
//
//		require.Equal(t, View(0), term.GetView(), "Term should have view=0 on init")
//		net.TriggerElection()
//		require.Equal(t, View(1), term.GetView(), "Term should have view=1 after one election")
//		nvm := builders.ANewViewMessage(node.KeyManager, height, view, ppmContentBuilder, nil, block)
//		term.OnReceiveNewView(ctx, nvm)
//		require.Equal(t, View(1), term.GetView(), "Term should have view=1")
//	})
//}
