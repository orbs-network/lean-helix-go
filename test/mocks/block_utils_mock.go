// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
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
	RequestNewBlockSns     *test.Sns

	ValidationCounter int
	ValidationSns     *test.Sns
	PauseOnValidation bool
	ValidationResult  bool
}

func NewMockBlockUtils(blocksPool *BlocksPool) *MockBlockUtils {
	return &MockBlockUtils{
		blocksPool: blocksPool,

		PauseOnRequestNewBlock: false,
		RequestNewBlockSns:     test.NewSignalAndStop(),

		ValidationCounter: 0,
		ValidationSns:     test.NewSignalAndStop(),
		PauseOnValidation: false,
		ValidationResult:  true,
	}
}

func (b *MockBlockUtils) RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
	if b.PauseOnRequestNewBlock {
		b.RequestNewBlockSns.SignalAndStop(ctx)
	}

	block := b.blocksPool.PopBlock(prevBlock)
	blockHash := CalculateBlockHash(block)
	return block, blockHash
}

func (b *MockBlockUtils) ValidateBlockCommitment(blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash) bool {
	return CalculateBlockHash(block).Equal(blockHash)
}

func (b *MockBlockUtils) CounterOfValidation() int {
	return b.ValidationCounter
}

func (b *MockBlockUtils) ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash, prevBlock interfaces.Block) error {
	b.ValidationCounter++
	if b.PauseOnValidation {
		b.ValidationSns.SignalAndStop(ctx)
	}

	if !b.ValidationResult {
		return errors.New("some errors")
	}
	return nil
}
