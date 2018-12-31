package termincommittee

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockproof"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodeSync(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.CreateBlock(interfaces.GenesisBlock)
		block2 := mocks.CreateBlock(block1)
		block3 := mocks.CreateBlock(block2)

		net := network.ATestNetwork(4, block1, block2, block3)
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
		require.True(t, node3.GetLatestBlock() == interfaces.GenesisBlock)

		net.WaitForNodeToRequestNewBlock(node0)

		// opening node3's network to messages
		node3.Gossip.ClearIncomingWhitelist()

		// syncing node3
		latestBlock := node0.GetLatestBlock()
		latestBlockProof := node0.GetLatestBlockProof()
		node3.Sync(ctx, latestBlock, latestBlockProof)

		net.ResumeNodeRequestNewBlock(node0)

		// now that node3 is synced, they all should progress to block3
		net.WaitForNodesToCommitASpecificBlock(block3, node0, node1, node2, node3)
	})
}

func TestAValidBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.CreateBlock(interfaces.GenesisBlock)
		block2 := mocks.CreateBlock(block1)
		block3 := mocks.CreateBlock(block2)

		net := network.ABasicTestNetwork()

		net.StartConsensus(ctx)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		cm0 := builders.ACommitMessage(node1.KeyManager, node1.MemberId, block3.Height(), 6, block3)
		cm1 := builders.ACommitMessage(node2.KeyManager, node2.MemberId, block3.Height(), 6, block3)
		cm2 := builders.ACommitMessage(node3.KeyManager, node3.MemberId, block3.Height(), 6, block3)

		commitMessages := []*interfaces.CommitMessage{cm0, cm1, cm2}

		blockProof := blockproof.GenerateLeanHelixBlockProof(commitMessages).Raw()
		require.False(t, node0.ValidateBlockConsensus(ctx, nil, blockProof))
		require.True(t, node0.ValidateBlockConsensus(ctx, block3, blockProof))
	})
}

func TestThatWeDoNotAcceptNilBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()

		net.StartConsensus(ctx)

		block1 := mocks.CreateBlock(interfaces.GenesisBlock)
		block2 := mocks.CreateBlock(block1)
		block3 := mocks.CreateBlock(block2)
		require.False(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, nil))
		require.False(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, []byte{}))
	})
}

func TestThatBlockRefInsideProofValidation(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		net.StartConsensus(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]

		block1 := mocks.CreateBlock(interfaces.GenesisBlock)
		block2 := mocks.CreateBlock(block1)
		block3 := mocks.CreateBlock(block2)
		blockHeight := block3.Height()

		goodBlockRef := generateACommitBlockRefBuilder(blockHeight, block3)
		signatures := generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2)

		nilBlockRefProof := (&protocol.BlockProofBuilder{
			BlockRef: nil,
			Nodes:    signatures,
		}).Build()

		badBlockHeightProof := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_COMMIT,
				BlockHeight: 666,
				BlockHash:   mocks.CalculateBlockHash(block3),
			},
			Nodes: signatures,
		}).Build()

		badMessageTypeProof := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_NEW_VIEW,
				BlockHeight: blockHeight,
				BlockHash:   mocks.CalculateBlockHash(block3),
			},
			Nodes: signatures,
		}).Build()

		badBlockHash := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_COMMIT,
				BlockHeight: blockHeight,
				BlockHash:   mocks.CalculateBlockHash(block1),
			},
			Nodes: signatures,
		}).Build()

		goodProof := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_COMMIT,
				BlockHeight: blockHeight,
				BlockHash:   mocks.CalculateBlockHash(block3),
			},
			Nodes: signatures,
		}).Build()

		require.True(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, nilBlockRefProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badBlockHeightProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badMessageTypeProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badBlockHash.Raw()))
	})
}

func generateACommitBlockRefBuilder(blockHeight primitives.BlockHeight, block interfaces.Block) *protocol.BlockRefBuilder {
	return &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_COMMIT,
		BlockHeight: blockHeight,
		BlockHash:   mocks.CalculateBlockHash(block),
	}
}

func generateSignatures(blockHeight primitives.BlockHeight, blockRef *protocol.BlockRef, nodes ...*network.Node) []*protocol.SenderSignatureBuilder {
	var result []*protocol.SenderSignatureBuilder
	for _, node := range nodes {
		result = append(result, &protocol.SenderSignatureBuilder{
			MemberId:  node.MemberId,
			Signature: node.KeyManager.SignConsensusMessage(blockHeight, blockRef.Raw()),
		})
	}

	return result
}

func TestCommitsWhenValidatingBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		outOfNetworkNode := network.ADummyNode()

		net.StartConsensus(ctx)

		block1 := mocks.CreateBlock(interfaces.GenesisBlock)
		block2 := mocks.CreateBlock(block1)
		block3 := mocks.CreateBlock(block2)

		blockHeight := block3.Height()
		goodBlockRef := generateACommitBlockRefBuilder(blockHeight, block3)

		// good proof
		goodProof := &protocol.BlockProofBuilder{
			BlockRef: goodBlockRef,
			Nodes:    generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
		}

		// proof with bad block height
		badBlockRefBlockHeightProof := &protocol.BlockProofBuilder{
			BlockRef: generateACommitBlockRefBuilder(666, block3),
			Nodes:    generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
		}

		// proof with not enough nodes
		noQuorumProof := &protocol.BlockProofBuilder{
			BlockRef: goodBlockRef,
			Nodes:    generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1),
		}

		// proof with duplicate nodes
		duplicateNodesProof := &protocol.BlockProofBuilder{
			BlockRef: goodBlockRef,
			Nodes:    generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node1),
		}

		// proof with a node that's not part of the network
		unknownNodeProof := &protocol.BlockProofBuilder{
			BlockRef: goodBlockRef,
			Nodes:    generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, outOfNetworkNode),
		}

		require.True(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, noQuorumProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badBlockRefBlockHeightProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, duplicateNodesProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, unknownNodeProof.Build().Raw()))
	})
}
