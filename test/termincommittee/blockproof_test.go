package termincommittee

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodeSync(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(leanhelix.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block2)

		net := builders.ATestNetwork(4, block1, block2, block3)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		net.NodesPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		// closing node3's network to messages (To make it out of sync)
		node3.Gossip.SetIncomingWhitelist([]primitives.MemberId{})

		// node0, node1, and node2 are progressing to block2
		net.WaitForNodeToRequestNewBlock(node0)
		net.ResumeNodeRequestNewBlock(node0)
		net.WaitForNodesToCommitASpecificBlock(block1, node0, node1, node2)

		net.WaitForNodeToRequestNewBlock(node0)
		net.ResumeNodeRequestNewBlock(node0)
		net.WaitForNodesToCommitASpecificBlock(block2, node0, node1, node2)

		// node3 is still "stuck" on the genesis block
		require.True(t, node3.GetLatestBlock() == leanhelix.GenesisBlock)

		net.WaitForNodeToRequestNewBlock(node0)

		// opening node3's network to messages
		node3.Gossip.ClearIncomingWhitelist()

		// syncing node3
		latestBlock := node0.GetLatestBlock()
		latestBlockProof := node0.GetLatestBlockProof()
		node3.Sync(latestBlock, latestBlockProof)

		net.ResumeNodeRequestNewBlock(node0)

		// now that node3 is synced, they all should progress to block3
		net.WaitForNodesToCommitASpecificBlock(block3, node0, node1, node2, node3)
	})
}

func TestAValidBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := builders.CreateBlock(leanhelix.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block2)

		net := builders.ABasicTestNetwork()

		net.StartConsensus(ctx)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		cm0 := builders.ACommitMessage(node1.KeyManager, node1.MemberId, block3.Height(), 6, block3)
		cm1 := builders.ACommitMessage(node2.KeyManager, node2.MemberId, block3.Height(), 6, block3)
		cm2 := builders.ACommitMessage(node3.KeyManager, node3.MemberId, block3.Height(), 6, block3)

		commitMessages := []*leanhelix.CommitMessage{cm0, cm1, cm2}

		blockProof := leanhelix.GenerateLeanHelixBlockProof(commitMessages).Raw()
		require.True(t, node0.ValidateBlockConsensus(block3, blockProof))
	})
}

func TestThatWeDoNotAcceptNilBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork()

		net.StartConsensus(ctx)

		block1 := builders.CreateBlock(leanhelix.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block2)
		require.False(t, net.Nodes[0].ValidateBlockConsensus(block3, nil))
		require.False(t, net.Nodes[0].ValidateBlockConsensus(block3, []byte{}))
	})
}

func TestThatABlockProofMatchTheGivenBlockHeight(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := builders.ABasicTestNetwork()

		net.StartConsensus(ctx)

		block1 := builders.CreateBlock(leanhelix.GenesisBlock)
		block2 := builders.CreateBlock(block1)
		block3 := builders.CreateBlock(block2)

		blockRef := protocol.BlockRefBuilder{BlockHeight: 666}
		proof := (&protocol.BlockProofBuilder{
			BlockRef: &blockRef,
		}).Build().Raw()
		require.False(t, net.Nodes[0].ValidateBlockConsensus(block3, proof))
		require.False(t, net.Nodes[0].ValidateBlockConsensus(nil, proof))
	})
}
