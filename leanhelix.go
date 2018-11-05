package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type leanHelix struct {
	config *Config
}

func (lh *leanHelix) ValidateBlockConsensus(block Block, blockProof *BlockProof, prevBlockProof *BlockProof) {
	panic("impl me")
}

func (lh *leanHelix) RegisterOnCommitted(cb func(block Block)) {
	// TODO: implement
}

func (lh *leanHelix) Dispose() {
	// TODO: implement
}

func (lh *leanHelix) Start(parentCtx context.Context, blockHeight primitives.BlockHeight) {
}

func (lh *leanHelix) IsLeader() bool {
	// TODO: implement
	return false
}
