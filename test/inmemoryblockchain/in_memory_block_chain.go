package inmemoryblockchain

import "github.com/orbs-network/lean-helix-go/types"

type InMemoryBlockChain struct {
	blockChain []*types.Block
}

func NewInMemoryBlockChain() *InMemoryBlockChain {
	return &InMemoryBlockChain{
		blockChain: []*types.Block{GenesisBlock},
	}
}

func (bs *InMemoryBlockChain) AppendBlockToChain(block *types.Block) {
	bs.blockChain = append(bs.blockChain, block)
}

func (bs *InMemoryBlockChain) GetLastBlock() *types.Block {
	return bs.blockChain[len(bs.blockChain)-1]
}
