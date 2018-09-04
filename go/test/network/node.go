package network

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/test/inmemoryblockchain"
)

type Node struct {
	PublicKey  lh.PublicKey
	Config     *lh.Config
	pbft       *lh.PBFT
	blockChain *inmemoryblockchain.InMemoryBlockChain
}

func NewNode(publicKey lh.PublicKey, config *lh.Config) *Node {
	pbft := lh.NewPBFT(config)
	node := &Node{
		PublicKey:  publicKey,
		Config:     config,
		pbft:       pbft,
		blockChain: inmemoryblockchain.NewInMemoryBlockChain(),
	}
	pbft.RegisterOnCommitted(node.onCommittedBlock)
	return node
}

func (node *Node) GetLatestCommittedBlock() *lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) IsLeader() bool {
	return node.pbft.IsLeader()
}

func (node *Node) TriggerElection() {
	// TODO fix error
	// node.Config.ElectionTrigger.(electiontrigger.ElectionTriggerMock).Trigger()
}

func (node *Node) onCommittedBlock(block *lh.Block) {
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
