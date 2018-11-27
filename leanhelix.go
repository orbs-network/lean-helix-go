package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type leanHelix struct {
	config              *Config
	logger              Logger
	filter              *ConsensusMessageFilter
	subscriptionToken   int
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
	lh.logger.Debug("LeanHelix.Start")
	for {
		leanHelixTerm := NewLeanHelixTerm(ctx, lh.config, lh.filter, blockHeight)
		block := leanHelixTerm.WaitForBlock(ctx)
		lh.notifyCommitted(block)
		blockHeight++
	}
}

func (lh *leanHelix) Dispose() {
	lh.config.NetworkCommunication.UnregisterOnMessage(lh.subscriptionToken)
}

func NewLeanHelix(config *Config) LeanHelix {
	if config.Logger == nil {
		config.Logger = NewSilentLogger()
	}
	filter := NewConsensusMessageFilter(config.KeyManager.MyPublicKey(), config.Logger)
	subscriptionToken := config.NetworkCommunication.RegisterOnMessage(filter.OnGossipMessage)
	return &leanHelix{config, config.Logger, filter, subscriptionToken, nil}
}
