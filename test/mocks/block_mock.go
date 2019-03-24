// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

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

func ABlock(previousBlock interfaces.Block) interfaces.Block {
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
