package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	"github.com/orbs-network/lean-helix-go/primitives"
)

// TODO looks identical to Config, why is this needed?
type TermConfig struct {
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	ElectionTrigger      ElectionTrigger
	Storage              Storage
	Logger               log.BasicLogger
	//MessageFactory       *MessageFactory
}

type leanHelix struct {
	ctxCancel context.CancelFunc
	config    *Config
	log       log.BasicLogger
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

	ctx, ctxCancel := context.WithCancel(parentCtx)
	lh.ctxCancel = ctxCancel

	// TODO: create an infinite loop which can be stopped by context.Done()

	for {
		select {

		// case: some channel that fires when consensus completed successfully or with error
		case <-ctx.Done():
			lh.log.Info("Context.done called")
			lh.GracefulShutdown()

		}
	}

}

func (lh *leanHelix) IsLeader() bool {
	// TODO: implement
	return false
}

func (lh *leanHelix) GracefulShutdown() {
	lh.log.Info("LeanHelix.GracefulShutdown() called")
	lh.ctxCancel()
}

func BuildTermConfig(config *Config) *TermConfig {
	return &TermConfig{
		NetworkCommunication: config.NetworkCommunication,
		BlockUtils:           config.BlockUtils,
		KeyManager:           config.KeyManager,
		ElectionTrigger:      config.ElectionTrigger,
		Storage:              config.Storage, // TODO should this default to InMemoryStorage if nil??
		Logger:               config.Logger,
	}
}
