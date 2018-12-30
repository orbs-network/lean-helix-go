package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func CompareSenderSignature(commitMessage *leanhelix.CommitMessage, senderSignature *protocol.SenderSignature) bool {
	sender := commitMessage.Content().Sender()
	return sender.Signature().Equal(senderSignature.Signature()) && sender.MemberId().Equal(senderSignature.MemberId())
}

func TestGeneratingBlockProof(t *testing.T) {
	block := builders.CreateBlock(leanhelix.GenesisBlock)

	memberId0 := primitives.MemberId("Member0")
	memberId1 := primitives.MemberId("Member1")
	memberId2 := primitives.MemberId("Member2")
	memberId3 := primitives.MemberId("Member3")

	node0KeyManager := mocks.NewMockKeyManager(memberId0)
	node1KeyManager := mocks.NewMockKeyManager(memberId1)
	node2KeyManager := mocks.NewMockKeyManager(memberId2)
	node3KeyManager := mocks.NewMockKeyManager(memberId3)

	cm0 := builders.ACommitMessage(node1KeyManager, memberId1, 5, 6, block)
	cm1 := builders.ACommitMessage(node2KeyManager, memberId2, 5, 6, block)
	cm2 := builders.ACommitMessage(node3KeyManager, memberId3, 5, 6, block)
	cm3 := builders.ACommitMessage(node0KeyManager, memberId0, 5, 6, block)

	commitMessages := []*leanhelix.CommitMessage{cm0, cm1, cm2, cm3}

	blockProof := leanhelix.GenerateLeanHelixBlockProof(commitMessages)

	// BlockRef
	blockRef := blockProof.BlockRef()
	require.Equal(t, protocol.LEAN_HELIX_COMMIT, blockRef.MessageType())
	require.Equal(t, primitives.BlockHeight(5), blockRef.BlockHeight())
	require.Equal(t, primitives.View(6), blockRef.View())
	require.True(t, builders.CalculateBlockHash(block).Equal(blockRef.BlockHash()))

	i := blockProof.NodesIterator()
	require.True(t, CompareSenderSignature(cm0, i.NextNodes()))
	require.True(t, CompareSenderSignature(cm1, i.NextNodes()))
	require.True(t, CompareSenderSignature(cm2, i.NextNodes()))
	require.True(t, CompareSenderSignature(cm3, i.NextNodes()))
	require.False(t, i.HasNext())
}
