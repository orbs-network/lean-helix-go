package builders

import (
	"context"
	"crypto/sha256"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

func BlocksAreEqual(block1 lh.Block, block2 lh.Block) bool {
	return CalculateBlockHash(block1).Equal(CalculateBlockHash(block2))
}

func CalculateBlockHash(block lh.Block) Uint256 {
	mockBlock := block.(*MockBlock)
	str := fmt.Sprintf("%d_%s", mockBlock.Height(), mockBlock.Body())
	hash := sha256.Sum256([]byte(str))
	return hash[:]
}

type MockBlockUtils struct {
	upcomingBlocks    []lh.Block
	latestBlock       lh.Block
	validationCounter int
	ValidationSns     *Sns
}

func NewMockBlockUtils(upcomingBlocks []lh.Block) *MockBlockUtils {
	return &MockBlockUtils{
		upcomingBlocks:    upcomingBlocks,
		latestBlock:       GenesisBlock,
		validationCounter: 0,
		ValidationSns:     NewSignalAndStop(),
	}
}

func (b *MockBlockUtils) CalculateBlockHash(block lh.Block) Uint256 {
	return CalculateBlockHash(block)
}

func (b *MockBlockUtils) getNextBlock() lh.Block {
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

func (b *MockBlockUtils) RequestNewBlock(ctx context.Context, height BlockHeight) lh.Block {
	return b.getNextBlock()
}

func (b *MockBlockUtils) CounterOfValidation() int {
	return b.validationCounter
}
func (b *MockBlockUtils) ValidateBlock(block lh.Block) bool {
	b.validationCounter++
	b.ValidationSns.SignalAndStop()

	return true
}
