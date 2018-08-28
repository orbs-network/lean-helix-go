package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

var globalCounter = 0

var GenesisBlock = &lh.Block{
	Header: &lh.BlockHeader{
		Height:    0,
		BlockHash: lh.BlockHash("The Genesis Block"),
	},
	Body: []byte("The Genesis Block"),
}

func genBody() []byte {
	globalCounter++
	return []byte(fmt.Sprintf("Block %d", globalCounter))
}

func CreateBlock(previousBlock *lh.Block) *lh.Block {

	block := &lh.Block{
		Header: &lh.BlockHeader{
			Height:    previousBlock.Header.Height + 1,
			BlockHash: CalculateBlockHash(previousBlock),
		},
		Body: genBody(),
	}
	return block

}
