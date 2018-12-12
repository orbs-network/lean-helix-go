package byzantineattacks

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/orbs-network-go/test"
	"testing"
)

func TestThatWeReachConsensusWhere1OutOf4NodeIsByzantine(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := builders.CreateBlock(builders.GenesisBlock)
		net := builders.
			NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]leanhelix.Block{block}).
			Build()

		net.Nodes[3].Gossip.SetIncomingWhitelist([]primitives.Ed25519PublicKey{})

		net.StartConsensus(ctx)

		net.WaitForNodesToCommitABlock(net.Nodes[0], net.Nodes[1], net.Nodes[2])
	})
}

func TestThatWeReachConsensusWhere2OutOf7NodesAreByzantine(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block := builders.CreateBlock(builders.GenesisBlock)
		net := builders.
			NewTestNetworkBuilder().
			WithNodeCount(7).
			WithBlocks([]leanhelix.Block{block}).
			Build()

		net.Nodes[1].Gossip.SetIncomingWhitelist([]primitives.Ed25519PublicKey{})
		net.Nodes[2].Gossip.SetIncomingWhitelist([]primitives.Ed25519PublicKey{})

		net.StartConsensus(ctx)

		net.WaitForNodesToCommitABlock(net.Nodes[0], net.Nodes[3], net.Nodes[4], net.Nodes[5], net.Nodes[6])
	})
}
