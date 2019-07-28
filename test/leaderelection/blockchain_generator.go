package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"time"
)

func GenerateBlockChainFor(blocks []interfaces.Block) *mocks.InMemoryBlockchain {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h := NewStartedHarnessDontPauseOnRequestNewBlock(ctx, nil, true, blocks...)

	h.net.WaitUntilNodesCommitASpecificHeight(ctx, blocks[len(blocks)-1].Height())
	if ctx.Err() != nil {
		return nil
	}

	return h.net.Nodes[0].Blockchain().GetFirstXItems(len(blocks) + 1)
}
