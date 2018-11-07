package builders

import (
	"context"
	"crypto/sha256"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

func CalculateBlockHash(block lh.Block) Uint256 {
	mockBlock := block.(*MockBlock)
	str := fmt.Sprintf("%d_%s", mockBlock.Height(), mockBlock.Body())
	hash := sha256.Sum256([]byte(str))
	return hash[:]
}

type MockBlockUtils struct {
	upcomingBlocks    []lh.Block
	latestBlock       lh.Block
	autoValidate      bool
	pauseOnValidation bool
	validationCounter int
}

func NewMockBlockUtils(upcomingBlocks []lh.Block) *MockBlockUtils {
	return &MockBlockUtils{
		upcomingBlocks:    upcomingBlocks,
		latestBlock:       GenesisBlock,
		autoValidate:      true,
		pauseOnValidation: false,
		validationCounter: 0,
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

func (b *MockBlockUtils) ResolveAllValidations(isValid bool) {
}

func (b *MockBlockUtils) RequestCommittee() {
	panic("implement me")
}

func (b *MockBlockUtils) RequestNewBlock(ctx context.Context, height BlockHeight) lh.Block {
	return b.getNextBlock()
}

func (b *MockBlockUtils) CounterOfValidation() int {
	return b.validationCounter
}

func (b *MockBlockUtils) ValidateBlock(block lh.Block) bool {
	if b.autoValidate {
		b.validationCounter++
		return true
	}

	return false
}
