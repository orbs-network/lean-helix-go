package inmemoryblockchain

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

type InMemoryBlockChain struct {
	blockChain []*lh.Block
}

func NewInMemoryBlockChain() *InMemoryBlockChain {
	return &InMemoryBlockChain{
		blockChain: []*lh.Block{GenesisBlock},
	}
}

func (bs *InMemoryBlockChain) AppendBlockToChain(block *lh.Block) {
	bs.blockChain = append(bs.blockChain, block)
}

func (bs *InMemoryBlockChain) GetLastBlock() *lh.Block {
	return bs.blockChain[len(bs.blockChain)-1]
}
