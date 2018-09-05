package builders

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/types"
)

var globalCounter = 0

func genBody() []byte {
	globalCounter++
	return []byte(fmt.Sprintf("Block %d", globalCounter))
}

func CreateBlock(previousBlock *types.Block) *types.Block {

	block := &types.Block{
		Header: &types.BlockHeader{
			Height:    previousBlock.Header.Height + 1,
			BlockHash: leanhelix.CalculateBlockHash(previousBlock),
		},
		Body: genBody(),
	}
	return block

}
