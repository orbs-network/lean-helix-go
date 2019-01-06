package mocks

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type chainItem struct {
	block      interfaces.Block
	blockProof []byte
}
type InMemoryBlockChain struct {
	blockChain []*chainItem
}

func NewInMemoryBlockChain() *InMemoryBlockChain {
	return &InMemoryBlockChain{
		blockChain: []*chainItem{
			{interfaces.GenesisBlock, nil},
		},
	}
}

func (bs *InMemoryBlockChain) AppendBlockToChain(block interfaces.Block, blockProof []byte) {
	bs.blockChain = append(bs.blockChain, &chainItem{block, blockProof})
}

func (bs *InMemoryBlockChain) GetLastBlock() interfaces.Block {
	item := bs.blockChain[len(bs.blockChain)-1]
	return item.block
}

func (bs *InMemoryBlockChain) GetLastBlockProof() []byte {
	item := bs.blockChain[len(bs.blockChain)-1]
	return item.blockProof
}

func (bs *InMemoryBlockChain) GetBlockProofAt(height primitives.BlockHeight) []byte {
	item := bs.blockChain[height]
	return item.blockProof
}
