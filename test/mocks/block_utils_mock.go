package mocks

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/sns"
)

func CalculateBlockHash(block interfaces.Block) primitives.BlockHash {
	if block == interfaces.GenesisBlock {
		fmt.Printf("Genesis")
	}
	mockBlock := block.(*MockBlock)
	str := fmt.Sprintf("%d_%s", mockBlock.Height(), mockBlock.Body())
	hash := sha256.Sum256([]byte(str))
	return hash[:]
}

type MockBlockUtils struct {
	blocksPool *BlocksPool

	PauseOnRequestNewBlock bool
	RequestNewBlockSns     *sns.Sns

	ValidationCounter int
	ValidationSns     *sns.Sns
	PauseOnValidation bool
	ValidationResult  bool
}

func NewMockBlockUtils(blocksPool *BlocksPool) *MockBlockUtils {
	return &MockBlockUtils{
		blocksPool: blocksPool,

		PauseOnRequestNewBlock: false,
		RequestNewBlockSns:     sns.NewSignalAndStop(),

		ValidationCounter: 0,
		ValidationSns:     sns.NewSignalAndStop(),
		PauseOnValidation: false,
		ValidationResult:  true,
	}
}

func (b *MockBlockUtils) RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
	if b.PauseOnRequestNewBlock {
		b.RequestNewBlockSns.SignalAndStop()
	}

	block := b.blocksPool.PopBlock()
	blockHash := CalculateBlockHash(block)
	return block, blockHash
}

func (b *MockBlockUtils) ValidateBlockCommitment(blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash) bool {
	return CalculateBlockHash(block).Equal(blockHash)
}

func (b *MockBlockUtils) CounterOfValidation() int {
	return b.ValidationCounter
}

func (b *MockBlockUtils) ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash, prevBlock interfaces.Block) bool {
	b.ValidationCounter++
	if b.PauseOnValidation {
		b.ValidationSns.SignalAndStop()
	}

	return b.ValidationResult
}
