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

func (b *MockBlockUtils) CalculateBlockHash(block lh.Block) Uint256 {
	return CalculateBlockHash(block)
}

func (b *MockBlockUtils) RequestNewBlock(ctx context.Context, prevBlock lh.Block) lh.Block {
	if b.PauseOnRequestNewBlock {
		b.RequestNewBlockSns.SignalAndStop()
	}
	return b.blocksPool.PopBlock()
}

func (b *MockBlockUtils) CounterOfValidation() int {
	return b.validationCounter
}

func (b *MockBlockUtils) ValidateBlock(block lh.Block) bool {
	b.validationCounter++
	if b.PauseOnValidation {
		b.ValidationSns.SignalAndStop()
	}

	return b.ValidationResult
}
