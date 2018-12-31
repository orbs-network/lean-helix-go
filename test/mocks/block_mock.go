package mocks

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

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

func CreateBlock(previousBlock interfaces.Block) interfaces.Block {
	var prevBlockHeight primitives.BlockHeight
	if previousBlock == interfaces.GenesisBlock {
		prevBlockHeight = 0
	} else {
		prevBlockHeight = previousBlock.Height()
	}

	newBlockHeight := prevBlockHeight + 1
	block := &MockBlock{
		height: newBlockHeight,
		body:   genBody(newBlockHeight),
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
