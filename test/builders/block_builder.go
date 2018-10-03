package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
)

var GenesisBlock = &block{
	term:      0,
	blockHash: lh.BlockHash("The Genesis Block"),
}

func (b *block) GetTerm() lh.BlockHeight {
	return b.term
}

func (h *block) GetBlockHash() lh.BlockHash {
	return h.blockHash
}

// block
type block struct {
	term      lh.BlockHeight
	blockHash lh.BlockHash
	body      []byte
}

var globalCounter = 0

func genBody() []byte {
	globalCounter++
	return []byte(fmt.Sprintf("Block %d", globalCounter))
}

func CreateBlock(previousBlock lh.Block) lh.Block {

	block := &block{
		term:      previousBlock.GetTerm() + 1,
		blockHash: CalculateBlockHash(previousBlock),
		body:      genBody(),
	}
	return block
}
