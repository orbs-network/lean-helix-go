package test

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

func compareSenderSignature(commitMessage *interfaces.CommitMessage, senderSignature *protocol.SenderSignature) bool {
	sender := commitMessage.Content().Sender()
	return sender.Signature().Equal(senderSignature.Signature()) && sender.MemberId().Equal(senderSignature.MemberId())
}

func TestGeneratingBlockProof(t *testing.T) {
	block := mocks.ABlock(interfaces.GenesisBlock)

	memberId0 := primitives.MemberId("Member0")
	memberId1 := primitives.MemberId("Member1")
	memberId2 := primitives.MemberId("Member2")
	memberId3 := primitives.MemberId("Member3")

	node0KeyManager := mocks.NewMockKeyManager(memberId0)
	node1KeyManager := mocks.NewMockKeyManager(memberId1)
	node2KeyManager := mocks.NewMockKeyManager(memberId2)
	node3KeyManager := mocks.NewMockKeyManager(memberId3)

	cm0 := builders.ACommitMessage(node1KeyManager, memberId1, 5, 6, block, 0)
	cm1 := builders.ACommitMessage(node2KeyManager, memberId2, 5, 6, block, 0)
	cm2 := builders.ACommitMessage(node3KeyManager, memberId3, 5, 6, block, 0)
	cm3 := builders.ACommitMessage(node0KeyManager, memberId0, 5, 6, block, 0)

	commitMessages := []*interfaces.CommitMessage{cm0, cm1, cm2, cm3}

	blockProof := blockproof.GenerateLeanHelixBlockProof(commitMessages)

	// BlockRef
	blockRef := blockProof.BlockRef()
	require.Equal(t, protocol.LEAN_HELIX_COMMIT, blockRef.MessageType())
	require.Equal(t, primitives.BlockHeight(5), blockRef.BlockHeight())
	require.Equal(t, primitives.View(6), blockRef.View())
	require.True(t, mocks.CalculateBlockHash(block).Equal(blockRef.BlockHash()))

	i := blockProof.NodesIterator()
	require.True(t, compareSenderSignature(cm0, i.NextNodes()))
	require.True(t, compareSenderSignature(cm1, i.NextNodes()))
	require.True(t, compareSenderSignature(cm2, i.NextNodes()))
	require.True(t, compareSenderSignature(cm3, i.NextNodes()))
	require.False(t, i.HasNext())
}

func TestNodeSync(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.ATestNetwork(4, block1, block2, block3)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		net.NodesPauseOnRequestNewBlock()
		net.StartConsensus(ctx)

		// closing node3's network to messages (To make it out of sync)
		node3.Communication.SetIncomingWhitelist([]primitives.MemberId{})

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
		node3.Communication.ClearIncomingWhitelist()

		// syncing node3
		latestBlock := node0.GetLatestBlock()
		latestBlockProof := node0.GetLatestBlockProof()
		prevBlockProof := node0.GetBlockProofAt(latestBlock.Height() - 1)
		node3.Sync(ctx, latestBlock, latestBlockProof, prevBlockProof)

		net.ResumeNodeRequestNewBlock(node0)

		// now that node3 is synced, they all should progress to block3
		net.WaitForNodesToCommitASpecificBlock(block3, node0, node1, node2, node3)
	})
}

func genBlockProofMessages(block interfaces.Block, view primitives.View, randomSeed uint64, nodes ...*network.Node) *protocol.BlockProof {
	var commitMessages []*interfaces.CommitMessage
	for _, node := range nodes {
		cm := builders.ACommitMessage(node.KeyManager, node.MemberId, block.Height(), view, block, randomSeed)
		commitMessages = append(commitMessages, cm)
	}

	return blockproof.GenerateLeanHelixBlockProof(commitMessages)
}

func TestAValidBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.ABasicTestNetwork()

		net.StartConsensus(ctx)
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		blockProof := genBlockProofMessages(block3, 6, 0, node1, node2, node3).Raw()
		prevBlockProof := genBlockProofMessages(block2, 3, 0, node1, node2, node3).Raw()
		require.True(t, node0.ValidateBlockConsensus(ctx, block3, blockProof, prevBlockProof))
		require.False(t, node0.ValidateBlockConsensus(ctx, nil, blockProof, prevBlockProof))
	})
}

func TestThatWeDoNotAcceptNilBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()

		net.StartConsensus(ctx)

		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)
		require.False(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, nil, nil))
		require.False(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, []byte{}, nil))
		require.False(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, nil, []byte{}))
		require.False(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, []byte{}, []byte{}))
	})
}

func TestThatBlockRefInsideProofValidation(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		net.StartConsensus(ctx)

		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]

		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)
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

		goodPrevProof := (&protocol.BlockProofBuilder{
			RandomSeedSignature: []byte{1, 2, 3},
		}).Build()

		goodProof := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_COMMIT,
				BlockHeight: blockHeight,
				BlockHash:   mocks.CalculateBlockHash(block3),
			},
			Nodes:               signatures,
			RandomSeedSignature: []byte{1, 2, 3},
		}).Build()

		require.True(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Raw(), goodPrevProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, nilBlockRefProof.Raw(), goodPrevProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badBlockHeightProof.Raw(), goodPrevProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badMessageTypeProof.Raw(), goodPrevProof.Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badBlockHash.Raw(), goodPrevProof.Raw()))
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

		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		blockHeight := block3.Height()
		goodBlockRef := generateACommitBlockRefBuilder(blockHeight, block3)

		// good prev proof
		goodPrevProof := &protocol.BlockProofBuilder{
			RandomSeedSignature: []byte{1, 2, 3},
		}

		// good proof
		goodProof := &protocol.BlockProofBuilder{
			BlockRef:            goodBlockRef,
			Nodes:               generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
			RandomSeedSignature: []byte{1, 2, 3},
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

		require.True(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Build().Raw(), goodPrevProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, noQuorumProof.Build().Raw(), goodPrevProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, badBlockRefBlockHeightProof.Build().Raw(), goodPrevProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, duplicateNodesProof.Build().Raw(), goodPrevProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, unknownNodeProof.Build().Raw(), goodPrevProof.Build().Raw()))
	})
}

func TestRandomSeedSignatureValidation(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		net := network.ABasicTestNetwork()
		node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]

		net.StartConsensus(ctx)

		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		blockHeight := block3.Height()
		goodBlockRef := generateACommitBlockRefBuilder(blockHeight, block3)

		// good prev proof
		goodPrevProof := &protocol.BlockProofBuilder{
			RandomSeedSignature: []byte{123},
		}

		// good proof
		goodProof := &protocol.BlockProofBuilder{
			BlockRef:            goodBlockRef,
			Nodes:               generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
			RandomSeedSignature: []byte{123},
		}

		// proof with no random seed signature
		noRSSProof := &protocol.BlockProofBuilder{
			BlockRef: goodBlockRef,
			Nodes:    generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
		}

		require.True(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Build().Raw(), goodPrevProof.Build().Raw()))
		require.False(t, node0.ValidateBlockConsensus(ctx, block3, noRSSProof.Build().Raw(), goodPrevProof.Build().Raw()))
	})
}
