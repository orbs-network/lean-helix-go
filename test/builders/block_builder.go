package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
)

var GenesisBlock = &block{
	height:    0,
	blockHash: lh.BlockHash("The Genesis Block"),
}

func (b *block) GetTerm() lh.BlockHeight {
	return b.height
}

func (h *block) GetBlockHash() lh.BlockHash {
	return h.blockHash
}

// block
type block struct {
	height    lh.BlockHeight
	blockHash lh.BlockHash
	body      []byte
}

func (b *block) GetHeight() lh.BlockHeight {
	return b.height
}

var globalCounter = 0

func genBody() []byte {
	globalCounter++
	return []byte(fmt.Sprintf("Block %d", globalCounter))
}

func CreateBlock(previousBlock lh.Block) lh.Block {

	block := &block{
		height:    previousBlock.GetHeight() + 1,
		blockHash: CalculateBlockHash(previousBlock),
		body:      genBody(),
	}
	return block
}
