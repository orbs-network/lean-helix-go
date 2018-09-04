package network

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/test/inmemoryblockchain"
)

type Node struct {
	PublicKey  lh.PublicKey
	blockChain *inmemoryblockchain.InMemoryBlockChain
}

func NewNode(publicKey lh.PublicKey) *Node {
	return &Node{
		PublicKey:  publicKey,
		blockChain: inmemoryblockchain.NewInMemoryBlockChain(),
	}
}

func (node *Node) GetLatestCommittedBlock() *lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) IsLeader() bool {
	// TODO: Implement
	return false
}

func (node *Node) TriggerElection() {
	// TODO: Implement
}

func (node *Node) onCommittedBlock() {
	// TODO: Implement
}

func (node *Node) StartConsensus() {
	// TODO: Implement
}

func (node *Node) Dispose() {
	// TODO: Implement
}
