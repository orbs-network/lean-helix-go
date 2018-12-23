package builders

import (
	"github.com/orbs-network/lean-helix-go"
)

type chainItem struct {
	block      leanhelix.Block
	blockProof []byte
}
type InMemoryBlockChain struct {
	blockChain []*chainItem
}

func NewInMemoryBlockChain() *InMemoryBlockChain {
	return &InMemoryBlockChain{
		blockChain: []*chainItem{
			{leanhelix.GenesisBlock, nil},
		},
	}
}

func (bs *InMemoryBlockChain) AppendBlockToChain(block leanhelix.Block, blockProof []byte) {
	bs.blockChain = append(bs.blockChain, &chainItem{block, blockProof})
}

func (bs *InMemoryBlockChain) GetLastBlock() leanhelix.Block {
	item := bs.blockChain[len(bs.blockChain)-1]
	return item.block
}

func (bs *InMemoryBlockChain) GetLastBlockProof() []byte {
	item := bs.blockChain[len(bs.blockChain)-1]
	return item.blockProof
}
