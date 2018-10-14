package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

var GenesisBlock = &MockBlock{
	height:    0,
	blockHash: Uint256("The Genesis Block"),
}

func CalculateBlockHash(block lh.Block) Uint256 {
	panic("impl me!")
}

func (b *MockBlock) GetTerm() BlockHeight {
	return b.height
}

func (h *MockBlock) GetBlockHash() Uint256 {
	return h.blockHash
}

// MockBlock
type MockBlock struct {
	height    BlockHeight
	blockHash Uint256
	body      []byte
}

func (b *MockBlock) GetHeight() BlockHeight {
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
