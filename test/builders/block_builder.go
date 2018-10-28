package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

var GenesisBlock = CreateBlock(nil)

// MockBlock
type MockBlock struct {
	height    BlockHeight
	blockHash Uint256
	body      string
}

func (b *MockBlock) BlockHash() Uint256 {
	return b.blockHash
}

func (b *MockBlock) Height() BlockHeight {
	return b.height
}

func (b *MockBlock) Body() string {
	return b.body
}

func CreateBlock(previousBlock lh.Block) lh.Block {
	var height BlockHeight = 0
	if previousBlock != nil {
		height = previousBlock.Height() + 1
	}

	block := &MockBlock{
		height: height,
		body:   genBody(height),
	}
	block.blockHash = CalculateBlockHash(block)
	return block
}

var blocksCounter = 0

func genBody(height BlockHeight) string {
	body := fmt.Sprintf("Block #%d Height:%d", blocksCounter, height)
	if height == 0 {
		body = body + " (Genesis)"
	}
	blocksCounter++
	return body
}
