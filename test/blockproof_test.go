package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"testing"
)

func TestGeneratingBlockProof(t *testing.T) {
	block := builders.CreateBlock(leanhelix.GenesisBlock)

	memberId0 := primitives.MemberId("Member0")
	memberId1 := primitives.MemberId("Member1")
	memberId2 := primitives.MemberId("Member2")
	memberId3 := primitives.MemberId("Member3")

	node0KeyManager := builders.NewMockKeyManager(memberId0)
	node1KeyManager := builders.NewMockKeyManager(memberId1)
	node2KeyManager := builders.NewMockKeyManager(memberId2)
	node3KeyManager := builders.NewMockKeyManager(memberId3)

	cm0 := builders.ACommitMessage(node1KeyManager, memberId1, 1, 1, block)
	cm1 := builders.ACommitMessage(node2KeyManager, memberId2, 1, 1, block)
	cm2 := builders.ACommitMessage(node3KeyManager, memberId3, 1, 1, block)
	cm3 := builders.ACommitMessage(node0KeyManager, memberId0, 1, 1, block)

	commitMessages := []*leanhelix.CommitMessage{cm0, cm1, cm2, cm3}

	aggregateFunc := func(blockHeight primitives.BlockHeight, randomSeedShares []*protocol.SenderSignature) primitives.RandomSeedSignature {
		return primitives.RandomSeedSignature("ResultRandomSeedSignature")
	}

	leanhelix.GenerateLeanHelixBlockProof(commitMessages, aggregateFunc)
}
