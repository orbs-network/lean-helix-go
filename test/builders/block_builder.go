package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
)

var GenesisBlock = &MockBlock{
	height:    0,
	blockHash: lh.BlockHash("The Genesis Block"),
}

func CalculateBlockHash(block lh.Block) lh.BlockHash {
	panic("impl me!")
}

func (b *MockBlock) GetTerm() lh.BlockHeight {
	return b.height
}

func (h *MockBlock) GetBlockHash() lh.BlockHash {
	return h.blockHash
}

// MockBlock
type MockBlock struct {
	height    lh.BlockHeight
	blockHash lh.BlockHash
	body      []byte
}

func (b *MockBlock) GetHeight() lh.BlockHeight {
	return b.height
}

var globalCounter = 0

func genBody() []byte {
	globalCounter++
	return []byte(fmt.Sprintf("Block %d", globalCounter))
}

func CreateBlock(previousBlock lh.Block) lh.Block {

	block := &MockBlock{
		height:    previousBlock.GetHeight() + 1,
		blockHash: CalculateBlockHash(previousBlock),
		body:      genBody(),
	}
	return block
}
