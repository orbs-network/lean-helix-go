package builders

import (
	"github.com/orbs-network/lean-helix-go"
)

type InMemoryBlockChain struct {
	blockChain []leanhelix.Block
}

func NewInMemoryBlockChain() *InMemoryBlockChain {
	return &InMemoryBlockChain{
		blockChain: []leanhelix.Block{GenesisBlock},
	}
}

func (bs *InMemoryBlockChain) AppendBlockToChain(block leanhelix.Block) {
	bs.blockChain = append(bs.blockChain, block)
}

func (bs *InMemoryBlockChain) GetLastBlock() leanhelix.Block {
	return bs.blockChain[len(bs.blockChain)-1]
}
