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

func (lh *leanHelix) Start(ctx context.Context, blockHeight primitives.BlockHeight) {
	filter := NewConsensusMessageFilter(lh.config.KeyManager.MyPublicKey())
	subscriptionToken := lh.config.NetworkCommunication.RegisterOnMessage(filter.OnGossipMessage)
	for {
		leanHelixTerm := NewLeanHelixTerm(lh.config, filter, blockHeight)
		block := leanHelixTerm.WaitForBlock(ctx)
		lh.notifyCommitted(block)
		blockHeight++
	}
	lh.config.NetworkCommunication.UnregisterOnMessage(subscriptionToken)
}

func NewLeanHelix(config *Config) LeanHelix {
	return &leanHelix{
		config: config,
	}
}
