package builders

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
)

type MockBlockUtils struct {
	mock.Mock
	upcomingBlocks []lh.Block
	latestBlock    lh.Block
}

func NewMockBlockUtils(upcomingBlocks []lh.Block) *MockBlockUtils {
	return &MockBlockUtils{
		upcomingBlocks: upcomingBlocks,
		latestBlock:    GenesisBlock,
	}
}

func CalculateBlockHash(b lh.Block) lh.BlockHash {
	testBlock := b.(*block)
	return lh.BlockHash(fmt.Sprintf("%s_%d_%s", testBlock.body, testBlock.GetTerm(), testBlock.GetBlockHash()))
}

func (b MockBlockUtils) CalculateBlockHash(block lh.Block) lh.BlockHash {
	return CalculateBlockHash(block)
}

func (b MockBlockUtils) ProvideNextBlock() lh.Block {

	var nextBlock lh.Block
	if len(b.upcomingBlocks) > 0 {
		// Simple queue impl, see https://github.com/golang/go/wiki/SliceTricks
		nextBlock, b.upcomingBlocks = b.upcomingBlocks[0], b.upcomingBlocks[1:]
	} else {
		nextBlock = CreateBlock(b.latestBlock)
	}
	b.latestBlock = nextBlock
	return nextBlock
}

func (b MockBlockUtils) ResolveAllValidations(isValid bool) {
}

func (b MockBlockUtils) RequestCommittee() {
	panic("implement me")
}

func (b MockBlockUtils) RequestNewBlock(height lh.BlockHeight) lh.Block {
	return b.ProvideNextBlock()
}

func (b MockBlockUtils) ValidateBlock() {
	panic("implement me")
}
