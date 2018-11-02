package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

// TODO looks identical to Config, why is this needed?
type TermConfig struct {
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	ElectionTrigger      ElectionTrigger
	Storage              Storage
	//messageFactory       *messageFactory
}

type leanHelix struct {
	ctxCancel context.CancelFunc
	config    *Config
}

func (lh *leanHelix) OnReceiveMessage(ctx context.Context, message ConsensusRawMessage) error {
	panic("implement me")
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
			lh.GracefulShutdown()

		}
	}

}

func (lh *leanHelix) IsLeader() bool {
	// TODO: implement
	return false
}

func (lh *leanHelix) GracefulShutdown() {
	lh.ctxCancel()
}

func BuildTermConfig(config *Config) *TermConfig {
	return &TermConfig{
		ElectionTrigger:      config.ElectionTrigger,
		NetworkCommunication: config.NetworkCommunication,
		Storage:              config.Storage, // TODO should this default to InMemoryStorage if nil??
		KeyManager:           config.KeyManager,
		BlockUtils:           config.BlockUtils,
	}
}
