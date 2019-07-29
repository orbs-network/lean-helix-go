package leaderelection

import (
	"github.com/orbs-network/lean-helix-go/services/blockproof"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
)

func GenerateProofsForTest(blocks []interfaces.Block, nodes []*network.Node) (*mocks.InMemoryBlockchain, error) {

	bc := mocks.NewInMemoryBlockchain()

	for _, b := range blocks {
		proof := generateProof(b, nodes)
		bc.AppendBlockToChain(b, proof.Raw())
	}

	return bc, nil

}

func generateProof(block interfaces.Block, nodes []*network.Node) *protocol.BlockProof {

	var instanceId primitives.InstanceId = 0
	commits := make([]*interfaces.CommitMessage, 0)

	for _, node := range nodes {
		commits = append(commits, builders.ACommitMessage(instanceId, node.KeyManager, node.MemberId, block.Height(), 0, block, 0))
	}
	blockProof := blockproof.GenerateLeanHelixBlockProof(nodes[0].KeyManager, commits)

	return blockProof
}
