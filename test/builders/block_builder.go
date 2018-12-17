package builders

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

var GenesisBlock = CreateBlock(nil)

// MockBlock
type MockBlock struct {
	height primitives.BlockHeight
	body   string
}

func (b *MockBlock) Height() primitives.BlockHeight {
	return b.height
}

func (b *MockBlock) Body() string {
	return b.body
}

func CreateBlock(previousBlock leanhelix.Block) leanhelix.Block {
	var height primitives.BlockHeight = 0
	if previousBlock != nil {
		height = previousBlock.Height() + 1
	}

	block := &MockBlock{
		height: height,
		body:   genBody(height),
	}
	return block
}

var blocksCounter = 0

func genBody(height primitives.BlockHeight) string {
	body := fmt.Sprintf("Block #%d Height:%d", blocksCounter, height)
	if height == 0 {
		body = body + " (Genesis)"
	}
	blocksCounter++
	return body
}
