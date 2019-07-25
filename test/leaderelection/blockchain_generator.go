package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

func GenerateBlockChainFor(blocks []interfaces.Block) *mocks.InMemoryBlockChain {

	ctx := context.Background()
	h := NewHarness(ctx, nil, false, blocks...)
	h.net.SetNodesToNotPauseOnRequestNewBlock()
	h.net.StartConsensus(ctx)
	h.net.WaitUntilNodesCommitASpecificBlock(ctx, blocks[len(blocks)-1])

	return h.net.Nodes[0].BlockChain()
}
