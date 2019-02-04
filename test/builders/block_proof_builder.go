package builders

import (
	"github.com/orbs-network/lean-helix-go/services/blockproof"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"math/rand"
)

func AMockBlockProof() *protocol.BlockProof {

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

	cm0 := ACommitMessage(instanceId, node0KeyManager, memberId0, 5, 6, block, 0)
	cm1 := ACommitMessage(instanceId, node1KeyManager, memberId1, 5, 6, block, 0)
	cm2 := ACommitMessage(instanceId, node2KeyManager, memberId2, 5, 6, block, 0)
	cm3 := ACommitMessage(instanceId, node3KeyManager, memberId3, 5, 6, block, 0)

	commitMessages := []*interfaces.CommitMessage{cm0, cm1, cm2, cm3}

	blockProof := blockproof.GenerateLeanHelixBlockProof(node1KeyManager, commitMessages)

	return blockProof

}
