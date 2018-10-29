package builders

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
)

type Node struct {
	Config     *lh.Config
	leanHelix  lh.LeanHelix
	blockChain *InMemoryBlockChain
}

func (node *Node) GetLatestCommittedBlock() lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) IsLeader() bool {
	return node.leanHelix.IsLeader()
}

func (node *Node) TriggerElection(ctx context.Context) {
	node.Config.ElectionTrigger.(*ElectionTriggerMock).Trigger(ctx)
}

func (node *Node) onCommittedBlock(block lh.Block) {
	node.blockChain.AppendBlockToChain(block)
}

func (node *Node) StartConsensus() {
	if node.leanHelix != nil {
		lastCommittedBlock := node.GetLatestCommittedBlock()
		node.leanHelix.Start(lastCommittedBlock.Height() + 1)
	}
}

func (node *Node) Dispose() {
	if node.leanHelix != nil {
		node.leanHelix.Dispose()
	}
}
