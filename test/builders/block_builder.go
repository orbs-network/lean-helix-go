package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
)

var GenesisBlock = &block{
	header: &blockHeader{
		term:      0,
		blockHash: lh.BlockHash("The Genesis Block"),
	},
	body: []byte("The Genesis Block"),
}

// BlockHeader
type blockHeader struct {
	term      lh.BlockHeight
	blockHash lh.BlockHash
}

func (h *blockHeader) Term() lh.BlockHeight {
	return h.term
}

func (h *blockHeader) BlockHash() lh.BlockHash {
	return h.blockHash
}

// block
type block struct {
	header *blockHeader
	body   []byte
}

func (b *block) Body() []byte {
	return b.body
}

func (b *block) Header() lh.BlockHeader {
	return b.header
}

var globalCounter = 0

func genBody() []byte {
	globalCounter++
	return []byte(fmt.Sprintf("Block %d", globalCounter))
}

func CreateBlock(previousBlock lh.Block) lh.Block {

	block := &block{
		header: &blockHeader{
			term:      previousBlock.Header().Term() + 1,
			blockHash: CalculateBlockHash(previousBlock),
		},
		body: genBody(),
	}
	return block
}
