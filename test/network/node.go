package network

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/inmemoryblockchain"
	"github.com/orbs-network/lean-helix-go/types"
)

type Node struct {
	PublicKey  types.PublicKey
	Config     *leanhelix.Config
	pbft       *leanhelix.LeanHelix
	blockChain *inmemoryblockchain.InMemoryBlockChain
}

func NewNode(publicKey types.PublicKey, config *leanhelix.Config) *Node {
	pbft := leanhelix.NewLeanHelix(config)
	node := &Node{
		PublicKey:  publicKey,
		Config:     config,
		pbft:       pbft,
		blockChain: inmemoryblockchain.NewInMemoryBlockChain(),
	}
	pbft.RegisterOnCommitted(node.onCommittedBlock)
	return node
}

func (node *Node) GetLatestCommittedBlock() *types.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) IsLeader() bool {
	return node.pbft.IsLeader()
}

func (node *Node) TriggerElection() {
	node.Config.ElectionTrigger.(*leanhelix.ElectionTriggerMock).Trigger()
}

func (node *Node) onCommittedBlock(block *types.Block) {
	node.blockChain.AppendBlockToChain(block)
}

func (node *Node) StartConsensus() {
	if node.pbft != nil {
		lastCommittedBlock := node.GetLatestCommittedBlock()
		node.pbft.Start(lastCommittedBlock.Header.Height)
	}
}

func (node *Node) Dispose() {
	if node.pbft != nil {
		node.pbft.Dispose()
	}
}
