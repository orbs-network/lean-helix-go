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
	"math"
)

func CalculateBlockHash(block interfaces.Block) primitives.BlockHash {
	if block == nil {
		return primitives.BlockHash{}
	}
	mockBlock := block.(*MockBlock)
	str := fmt.Sprintf("%d_%s", mockBlock.Height(), mockBlock.Body())
	hash := sha256.Sum256([]byte(str))
	return hash[:]
}

type PausableBlockUtils struct {
	interfaces.BlockUtils
	memberId   primitives.MemberId
	blocksPool *BlocksPool
	//PauseOnRequestNewBlock       bool
	RequestNewBlockCallsLeftUntilItPausesWhenCounterIsZero int64
	RequestNewBlockLatch                                   *test.Latch
	ValidationLatch                                        *test.Latch
	PauseOnValidateBlock                                   bool
	failBlockProposalValidations                           bool
}

func NewMockBlockUtils(memberId primitives.MemberId, blocksPool *BlocksPool, logger interfaces.Logger) *PausableBlockUtils {
	return &PausableBlockUtils{
		memberId:   memberId,
		blocksPool: blocksPool,
		RequestNewBlockCallsLeftUntilItPausesWhenCounterIsZero: math.MaxInt64, // no pause by default
		RequestNewBlockLatch: test.NewLatch(logger),

		ValidationLatch:      test.NewLatch(logger),
		PauseOnValidateBlock: false,
	}
}

func (b *PausableBlockUtils) WithFailingBlockProposalValidations() *PausableBlockUtils {
	b.failBlockProposalValidations = true
	return b
}

func (b *PausableBlockUtils) RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, _ primitives.MemberId, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
	if b.RequestNewBlockCallsLeftUntilItPausesWhenCounterIsZero == 0 {
		//fmt.Printf("ID=%s H=%d RequestNewBlockProposal: Sleeping until latch is resumed\n", b.memberId, blockHeight)
		b.RequestNewBlockLatch.WaitOnPauseThenWaitOnResume(ctx, b.memberId)
		//fmt.Printf("ID=%s H=%d RequestNewBlockProposal: Latch has resumed. ctx.Err: %v\n", b.memberId, blockHeight, ctx.Err())
	} else {
		b.RequestNewBlockCallsLeftUntilItPausesWhenCounterIsZero--
	}

	block := b.blocksPool.PopBlock(prevBlock)
	blockHash := CalculateBlockHash(block)
	return block, blockHash
}

func (b *PausableBlockUtils) ValidateBlockCommitment(blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash) bool {
	return CalculateBlockHash(block).Equal(blockHash)
}

func (b *PausableBlockUtils) ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, memberId primitives.MemberId, block interfaces.Block, blockHash primitives.BlockHash, prevBlock interfaces.Block) error {
	if b.PauseOnValidateBlock {
		b.ValidationLatch.WaitOnPauseThenWaitOnResume(ctx, b.memberId)
	}

	if b.failBlockProposalValidations {
		return errors.New("some errors")
	}
	return nil
}
