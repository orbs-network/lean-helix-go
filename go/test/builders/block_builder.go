package builders

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/go/block"
	"github.com/orbs-network/lean-helix-go/go/test/blockutils"
)

var globalCounter = 0

var GenesisBlock = &block.Block{
	Header: &block.BlockHeader{
		Height:        0,
		PrevBlockHash: []byte{0},
	},
	Body: "The Genesis Block",
}

func genBody() string {
	globalCounter++
	return fmt.Sprintf("Block %d", globalCounter)
}

func CreateBlock(previousBlock *block.Block) *block.Block {

	block := &block.Block{
		Header: &block.BlockHeader{
			Height:        previousBlock.Header.Height + 1,
			PrevBlockHash: blockutils.CalculateBlockHash(previousBlock),
		},
		Body: genBody(),
	}
	return block

}
