package builders

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

func BlocksAreEqual(block1 leanhelix.Block, block2 leanhelix.Block) bool {
	return CalculateBlockHash(block1).Equal(CalculateBlockHash(block2))
}

func CalculateBlockHash(block leanhelix.Block) primitives.BlockHash {
	mockBlock := block.(*MockBlock)
	str := fmt.Sprintf("%d_%s", mockBlock.Height(), mockBlock.Body())
	hash := sha256.Sum256([]byte(str))
	return hash[:]
}

type MockBlockUtils struct {
	blocksPool *BlocksPool

	PauseOnRequestNewBlock bool
	RequestNewBlockSns     *Sns

	validationCounter int
	ValidationSns     *Sns
	PauseOnValidation bool
	ValidationResult  bool
}

func NewMockBlockUtils(blocksPool *BlocksPool) *MockBlockUtils {
	return &MockBlockUtils{
		blocksPool: blocksPool,

		PauseOnRequestNewBlock: false,
		RequestNewBlockSns:     NewSignalAndStop(),

		validationCounter: 0,
		ValidationSns:     NewSignalAndStop(),
		PauseOnValidation: false,
		ValidationResult:  true,
	}
}

func (b *MockBlockUtils) CalculateBlockHash(block leanhelix.Block) primitives.BlockHash {
	return CalculateBlockHash(block)
}

func (b *MockBlockUtils) RequestNewBlock(ctx context.Context, prevBlock leanhelix.Block) leanhelix.Block {
	if b.PauseOnRequestNewBlock {
		b.RequestNewBlockSns.SignalAndStop()
	}
	return b.blocksPool.PopBlock()
}

func (b *MockBlockUtils) CounterOfValidation() int {
	return b.validationCounter
}

func (b *MockBlockUtils) ValidateBlock(block leanhelix.Block) bool {
	b.validationCounter++
	if b.PauseOnValidation {
		b.ValidationSns.SignalAndStop()
	}

	return b.ValidationResult
}
