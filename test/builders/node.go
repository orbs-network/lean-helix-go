package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
)

type Node struct {
	PublicKey  lh.PublicKey
	Config     lh.Config
	pbft       lh.LeanHelix
	blockChain *InMemoryBlockChain
}

func NewNode(publicKey lh.PublicKey, config lh.Config) *Node {
	pbft := lh.NewLeanHelix(config)
	node := &Node{
		PublicKey:  publicKey,
		Config:     config,
		pbft:       pbft,
		blockChain: NewInMemoryBlockChain(),
	}
	pbft.RegisterOnCommitted(node.onCommittedBlock)
	return node
}

func (node *Node) GetLatestCommittedBlock() lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) IsLeader() bool {
	return node.pbft.IsLeader()
}

func (node *Node) TriggerElection() {
	node.Config.ElectionTrigger().(*ElectionTriggerMock).Trigger()
}

func (node *Node) onCommittedBlock(block lh.Block) {
	node.blockChain.AppendBlockToChain(block)
}

func (node *Node) StartConsensus() {
	if node.pbft != nil {
		lastCommittedBlock := node.GetLatestCommittedBlock()
		node.pbft.Start(lastCommittedBlock.Header().Term())
	}
}

func (node *Node) Dispose() {
	if node.pbft != nil {
		node.pbft.Dispose()
	}
}
