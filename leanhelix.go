package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type leanHelix struct {
	config              *Config
	commitSubscriptions []func(block Block)
}

func (lh *leanHelix) notifyCommitted(block Block) {
	for _, subscription := range lh.commitSubscriptions {
		subscription(block)
	}
}

func (lh *leanHelix) RegisterOnCommitted(cb func(block Block)) {
	lh.commitSubscriptions = append(lh.commitSubscriptions, cb)
}

func (lh *leanHelix) ValidateBlockConsensus(block Block, blockProof *BlockProof, prevBlockProof *BlockProof) {
	panic("impl me")
}

func (lh *leanHelix) Start(parentCtx context.Context, blockHeight primitives.BlockHeight) {
	filter := NewConsensusMessageFilter(lh.config.KeyManager.MyPublicKey())
	for {
		leanHelixTerm := NewLeanHelixTerm(lh.config, filter, blockHeight)
		block := leanHelixTerm.WaitForBlock()
		lh.notifyCommitted(block)
		blockHeight++
	}
}

func NewLeanHelix(config *Config) LeanHelix {
	return &leanHelix{
		config: config,
	}
}
