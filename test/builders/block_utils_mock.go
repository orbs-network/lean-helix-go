package builders

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/orbs-network/go-mock"
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

func (b *MockBlockUtils) CalculateBlockHash(block lh.Block) Uint256 {
	b.Called(block)
	return CalculateBlockHash(block)
}

func (b *MockBlockUtils) ProvideNextBlock() lh.Block {
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
	b.Called(isValid)
}

func (b *MockBlockUtils) RequestCommittee() {
	b.Called()
	panic("implement me")
}

func (b *MockBlockUtils) RequestNewBlock(ctx context.Context, height BlockHeight) lh.Block {
	//b.Called(height)
	return b.ProvideNextBlock()
}

func (b *MockBlockUtils) ValidateBlock(block lh.Block) bool {
	b.Called(block)
	panic("implement me")
}
