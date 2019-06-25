// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

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
	"math/rand"
	"testing"
)

func compareSenderSignature(commitMessage *interfaces.CommitMessage, senderSignature *protocol.SenderSignature) bool {
	sender := commitMessage.Content().Sender()
	return sender.Signature().Equal(senderSignature.Signature()) && sender.MemberId().Equal(senderSignature.MemberId())
}

func TestGeneratingBlockProof(t *testing.T) {
	block := mocks.ABlock(interfaces.GenesisBlock)
	instanceId := primitives.InstanceId(rand.Uint64())

	memberId0 := primitives.MemberId("Member0")
	memberId1 := primitives.MemberId("Member1")
	memberId2 := primitives.MemberId("Member2")
	memberId3 := primitives.MemberId("Member3")

	node0KeyManager := mocks.NewMockKeyManager(memberId0)
	node1KeyManager := mocks.NewMockKeyManager(memberId1)
	node2KeyManager := mocks.NewMockKeyManager(memberId2)
	node3KeyManager := mocks.NewMockKeyManager(memberId3)

	cm0 := builders.ACommitMessage(instanceId, node0KeyManager, memberId0, 5, 6, block, 0)
	cm1 := builders.ACommitMessage(instanceId, node1KeyManager, memberId1, 5, 6, block, 0)
	cm2 := builders.ACommitMessage(instanceId, node2KeyManager, memberId2, 5, 6, block, 0)
	cm3 := builders.ACommitMessage(instanceId, node3KeyManager, memberId3, 5, 6, block, 0)

	commitMessages := []*interfaces.CommitMessage{cm0, cm1, cm2, cm3}

	blockProof := blockproof.GenerateLeanHelixBlockProof(node1KeyManager, commitMessages)

	// BlockRef
	blockRef := blockProof.BlockRef()
	require.Equal(t, protocol.LEAN_HELIX_COMMIT, blockRef.MessageType())
	require.Equal(t, primitives.BlockHeight(5), blockRef.BlockHeight())
	require.Equal(t, primitives.View(6), blockRef.View())
	require.True(t, mocks.CalculateBlockHash(block).Equal(blockRef.BlockHash()))

	// Nodes
	i := blockProof.NodesIterator()
	require.True(t, compareSenderSignature(cm0, i.NextNodes()))
	require.True(t, compareSenderSignature(cm1, i.NextNodes()))
	require.True(t, compareSenderSignature(cm2, i.NextNodes()))
	require.True(t, compareSenderSignature(cm3, i.NextNodes()))
	require.False(t, i.HasNext())

	// RandomSeedSignature
	cShares := []*protocol.SenderSignature{
		(&protocol.SenderSignatureBuilder{
			MemberId:  memberId0,
			Signature: primitives.Signature(cm0.Content().Share()),
		}).Build(),
		(&protocol.SenderSignatureBuilder{
			MemberId:  memberId1,
			Signature: primitives.Signature(cm1.Content().Share()),
		}).Build(),
		(&protocol.SenderSignatureBuilder{
			MemberId:  memberId2,
			Signature: primitives.Signature(cm2.Content().Share()),
		}).Build(),
		(&protocol.SenderSignatureBuilder{
			MemberId:  memberId3,
			Signature: primitives.Signature(cm3.Content().Share()),
		}).Build(),
	}
	randomSeedSignature := node1KeyManager.AggregateRandomSeed(5, cShares)

	require.Equal(t, randomSeedSignature, blockProof.RandomSeedSignature())
}

func genBlockProofMessages(instanceId primitives.InstanceId, block interfaces.Block, view primitives.View, randomSeed uint64, nodes ...*network.Node) *protocol.BlockProof {
	var commitMessages []*interfaces.CommitMessage
	for _, node := range nodes {
		cm := builders.ACommitMessage(instanceId, node.KeyManager, node.MemberId, block.Height(), view, block, randomSeed)
		commitMessages = append(commitMessages, cm)
	}

	return blockproof.GenerateLeanHelixBlockProof(nodes[0].KeyManager, commitMessages)
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

		blockProof := genBlockProofMessages(net.InstanceId, block3, 6, 0, node1, node2, node3).Raw()
		prevBlockProof := genBlockProofMessages(net.InstanceId, block2, 3, 0, node1, node2, node3).Raw()
		require.Nil(t, node0.ValidateBlockConsensus(ctx, block3, blockProof, prevBlockProof))
		require.Error(t, node0.ValidateBlockConsensus(ctx, nil, blockProof, prevBlockProof))
	})
}

func TestAValidBlockProofWithNilPrevBlockProof(t *testing.T) {
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

		blockProof := genBlockProofMessages(net.InstanceId, block3, 6, 0, node1, node2, node3).Raw()
		require.Nil(t, node0.ValidateBlockConsensus(ctx, block3, blockProof, nil))
	})
}

func TestThatWeDoNotAcceptNilBlockProof(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.ABasicTestNetwork()
		net.StartConsensus(ctx)

		//node0 := net.Nodes[0]
		node1 := net.Nodes[1]
		node2 := net.Nodes[2]
		node3 := net.Nodes[3]

		prevBlockProof := genBlockProofMessages(net.InstanceId, block2, 3, 0, node1, node2, node3).Raw()

		require.Error(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, nil, prevBlockProof))
		require.Error(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, []byte{}, prevBlockProof))
		require.Error(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, nil, []byte{}))
		require.Error(t, net.Nodes[0].ValidateBlockConsensus(ctx, block3, []byte{}, []byte{}))
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

		goodBlockRef := generateACommitBlockRefBuilder(net.InstanceId, blockHeight, block3)
		signatures := generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2)
		goodRSS := node0.KeyManager.AggregateRandomSeed(blockHeight, nil)

		nilBlockRefProof := (&protocol.BlockProofBuilder{
			BlockRef:            nil,
			Nodes:               signatures,
			RandomSeedSignature: goodRSS,
		}).Build()

		badBlockHeightProof := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_COMMIT,
				InstanceId:  net.InstanceId,
				BlockHeight: 666,
				BlockHash:   mocks.CalculateBlockHash(block3),
			},
			Nodes:               signatures,
			RandomSeedSignature: goodRSS,
		}).Build()

		badMessageTypeProof := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_NEW_VIEW,
				InstanceId:  net.InstanceId,
				BlockHeight: blockHeight,
				BlockHash:   mocks.CalculateBlockHash(block3),
			},
			Nodes:               signatures,
			RandomSeedSignature: goodRSS,
		}).Build()

		badBlockHash := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_COMMIT,
				InstanceId:  net.InstanceId,
				BlockHeight: blockHeight,
				BlockHash:   mocks.CalculateBlockHash(block1),
			},
			Nodes:               signatures,
			RandomSeedSignature: goodRSS,
		}).Build()

		goodPrevProof := (&protocol.BlockProofBuilder{
			RandomSeedSignature: []byte{1, 2, 3},
		}).Build()

		goodProof := (&protocol.BlockProofBuilder{
			BlockRef: &protocol.BlockRefBuilder{
				MessageType: protocol.LEAN_HELIX_COMMIT,
				InstanceId:  net.InstanceId,
				BlockHeight: blockHeight,
				BlockHash:   mocks.CalculateBlockHash(block3),
			},
			Nodes:               signatures,
			RandomSeedSignature: goodRSS,
		}).Build()

		require.Nil(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Raw(), goodPrevProof.Raw()))
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, nilBlockRefProof.Raw(), goodPrevProof.Raw()))
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, badBlockHeightProof.Raw(), goodPrevProof.Raw()))
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, badMessageTypeProof.Raw(), goodPrevProof.Raw()))
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, badBlockHash.Raw(), goodPrevProof.Raw()))
	})
}

func generateACommitBlockRefBuilder(instanceId primitives.InstanceId, blockHeight primitives.BlockHeight, block interfaces.Block) *protocol.BlockRefBuilder {
	return &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_COMMIT,
		InstanceId:  instanceId,
		BlockHeight: blockHeight,
		BlockHash:   mocks.CalculateBlockHash(block),
	}
}

func generateSignatures(blockHeight primitives.BlockHeight, blockRef *protocol.BlockRef, nodes ...*network.Node) []*protocol.SenderSignatureBuilder {
	var result []*protocol.SenderSignatureBuilder
	for _, node := range nodes {
		result = append(result, &protocol.SenderSignatureBuilder{
			MemberId:  node.MemberId,
			Signature: node.KeyManager.SignConsensusMessage(context.Background(), blockHeight, blockRef.Raw()),
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
		goodBlockRef := generateACommitBlockRefBuilder(net.InstanceId, blockHeight, block3)

		goodRSS := node0.KeyManager.AggregateRandomSeed(blockHeight, nil)

		// good prev proof
		goodPrevProof := &protocol.BlockProofBuilder{
			RandomSeedSignature: []byte{1, 2, 3},
		}

		badInstanceId := primitives.InstanceId(888888)

		// good proof
		goodProof := &protocol.BlockProofBuilder{
			BlockRef:            goodBlockRef,
			Nodes:               generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
			RandomSeedSignature: goodRSS,
		}

		// proof with bad instance ID
		blockRefWithBadBlock := generateACommitBlockRefBuilder(badInstanceId, blockHeight, block3)
		badBlockRefInstanceIdProof := &protocol.BlockProofBuilder{
			BlockRef:            blockRefWithBadBlock,
			Nodes:               generateSignatures(blockHeight, blockRefWithBadBlock.Build(), node0, node1, node2),
			RandomSeedSignature: goodRSS,
		}

		// proof with bad block height
		blockRefWithBadInstanceId := generateACommitBlockRefBuilder(net.InstanceId, 666, block3)
		badBlockRefBlockHeightProof := &protocol.BlockProofBuilder{
			BlockRef:            blockRefWithBadInstanceId,
			Nodes:               generateSignatures(blockHeight, blockRefWithBadInstanceId.Build(), node0, node1, node2),
			RandomSeedSignature: goodRSS,
		}

		// proof with not enough nodes
		noQuorumProof := &protocol.BlockProofBuilder{
			BlockRef:            goodBlockRef,
			Nodes:               generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1),
			RandomSeedSignature: goodRSS,
		}

		// proof with duplicate nodes
		duplicateNodesProof := &protocol.BlockProofBuilder{
			BlockRef:            goodBlockRef,
			Nodes:               generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node1),
			RandomSeedSignature: goodRSS,
		}

		// proof with a node that's not part of the network
		unknownNodeProof := &protocol.BlockProofBuilder{
			BlockRef:            goodBlockRef,
			Nodes:               generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, outOfNetworkNode),
			RandomSeedSignature: goodRSS,
		}

		require.Nil(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Build().Raw(), goodPrevProof.Build().Raw()), "should succeed with good proof")
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, noQuorumProof.Build().Raw(), goodPrevProof.Build().Raw()), "should fail on not enough nodes in proof")
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, badBlockRefBlockHeightProof.Build().Raw(), goodPrevProof.Build().Raw()), "should fail on bad block height in proof")
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, badBlockRefInstanceIdProof.Build().Raw(), goodPrevProof.Build().Raw()), "should fail on bad instance ID")
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, duplicateNodesProof.Build().Raw(), goodPrevProof.Build().Raw()), "should fail on duplicate nodes in proof")
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, unknownNodeProof.Build().Raw(), goodPrevProof.Build().Raw()), "should fail on unknown node in proof")
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
		goodBlockRef := generateACommitBlockRefBuilder(net.InstanceId, blockHeight, block3)
		goodRSS := node0.KeyManager.AggregateRandomSeed(blockHeight, nil)

		// good prev proof
		goodPrevProof := &protocol.BlockProofBuilder{
			RandomSeedSignature: []byte{123},
		}

		// good proof
		goodProof := &protocol.BlockProofBuilder{
			BlockRef:            goodBlockRef,
			Nodes:               generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
			RandomSeedSignature: goodRSS,
		}

		// proof with no random seed signature
		noRSSProof := &protocol.BlockProofBuilder{
			BlockRef: goodBlockRef,
			Nodes:    generateSignatures(blockHeight, goodBlockRef.Build(), node0, node1, node2),
		}

		require.Nil(t, node0.ValidateBlockConsensus(ctx, block3, goodProof.Build().Raw(), goodPrevProof.Build().Raw()))
		require.Error(t, node0.ValidateBlockConsensus(ctx, block3, noRSSProof.Build().Raw(), goodPrevProof.Build().Raw()))
	})
}
