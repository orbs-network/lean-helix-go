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

type MockBlockUtils interface {
	interfaces.BlockUtils
	GetValidationResult() bool
	SetValidationResult(bool)
}

type PausableBlockUtils struct {
	MockBlockUtils
	blocksPool             *BlocksPool
	PauseOnRequestNewBlock bool
	RequestNewBlockLatch   *test.Latch
	ValidationLatch        *test.Latch
	PauseOnValidateBlock   bool
	ValidationResult       bool
}

func (b *PausableBlockUtils) GetValidationResult() bool {
	return b.ValidationResult
}

func (b *PausableBlockUtils) SetValidationResult(v bool) {
	b.ValidationResult = v
}

func NewMockBlockUtils(blocksPool *BlocksPool) *PausableBlockUtils {
	return &PausableBlockUtils{
		blocksPool:             blocksPool,
		PauseOnRequestNewBlock: false,
		RequestNewBlockLatch:   test.NewLatch(),

		ValidationLatch:      test.NewLatch(),
		PauseOnValidateBlock: false,
		ValidationResult:     true,
	}
}

func (b *PausableBlockUtils) RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
	if b.PauseOnRequestNewBlock {
		b.RequestNewBlockLatch.ReturnWhenLatchIsResumed(ctx)
	}

	block := b.blocksPool.PopBlock(prevBlock)
	blockHash := CalculateBlockHash(block)
	return block, blockHash
}

func (b *PausableBlockUtils) ValidateBlockCommitment(blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash) bool {
	return CalculateBlockHash(block).Equal(blockHash)
}

func (b *PausableBlockUtils) ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash, prevBlock interfaces.Block) error {
	if b.PauseOnValidateBlock {
		b.ValidationLatch.ReturnWhenLatchIsResumed(ctx)
	}

	if !b.ValidationResult {
		return errors.New("some errors")
	}
	return nil
}
