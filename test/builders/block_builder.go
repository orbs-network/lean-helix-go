package builders

import (
	"crypto/sha256"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

var GenesisBlock = &MockBlock{
	height:        0,
	blockHash:     Uint256("The Genesis Block"),
	prevBlockHash: nil,
}

func CalculateBlockHash(block lh.Block) Uint256 {
	hash := sha256.Sum256(block.PrevBlockHash())
	return hash[:]
}

// MockBlock
type MockBlock struct {
	height        BlockHeight
	blockHash     Uint256
	prevBlockHash Uint256
	body          []byte
}

func (b *MockBlock) PrevBlockHash() Uint256 {
	return b.prevBlockHash
}

func (b *MockBlock) BlockHash() Uint256 {
	return b.blockHash
}

func (b *MockBlock) Height() BlockHeight {
	return b.height
}

var globalCounter = 0

func CreateBlock(previousBlock lh.Block) lh.Block {

	block := &MockBlock{
		height:    previousBlock.Height() + 1,
		blockHash: CalculateBlockHash(previousBlock),
		body:      genBody(),
	}
	return block
}

func genBody() []byte {
	globalCounter++
	return []byte(fmt.Sprintf("Block %d", globalCounter))
}
