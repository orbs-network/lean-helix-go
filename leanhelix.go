package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type leanHelix struct {
	config              *Config
	leanHelixTerm       *leanHelixTerm
	messagesFilter      *NetworkMessageFilter
	commitSubscriptions []func(block Block)
}

func (lh *leanHelix) notifyCommitted(block Block) {
	for _, subscription := range lh.commitSubscriptions {
		subscription(block)
	}
}

func (lh *leanHelix) disposeLeanHelixTerm() {
	if lh.leanHelixTerm != nil {
		lh.leanHelixTerm.Dispose()
		lh.leanHelixTerm = nil
	}
}

func (lh *leanHelix) createLeanHelixTerm(blockHeight primitives.BlockHeight) {
	lh.leanHelixTerm = NewLeanHelixTerm(context.Background(), lh.config, blockHeight, func(block Block) {
		lh.notifyCommitted(block)
		lh.Start(context.Background(), block.Height()+1)
	})
	lh.messagesFilter.SetBlockHeight(context.Background(), blockHeight, lh.leanHelixTerm)
}

func (lh *leanHelix) IsLeader() bool {
	if lh.leanHelixTerm != nil {
		return lh.leanHelixTerm.IsLeader()
	}
	return false
}

func (lh *leanHelix) RegisterOnCommitted(cb func(block Block)) {
	lh.commitSubscriptions = append(lh.commitSubscriptions, cb)
}

func (lh *leanHelix) ValidateBlockConsensus(block Block, blockProof *BlockProof, prevBlockProof *BlockProof) {
	panic("impl me")
}

func (lh *leanHelix) Start(parentCtx context.Context, blockHeight primitives.BlockHeight) {
	go func() {
		lh.disposeLeanHelixTerm()
		lh.createLeanHelixTerm(blockHeight)
	}()
}

func (lh *leanHelix) Dispose() {
	lh.disposeLeanHelixTerm()
}

func NewLeanHelix(config *Config) LeanHelix {
	return &leanHelix{
		config:         config,
		messagesFilter: NewNetworkMessageFilter(config.NetworkCommunication, config.KeyManager.MyPublicKey()),
	}
}
