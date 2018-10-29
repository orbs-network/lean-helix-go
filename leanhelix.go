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
	Logger               log.BasicLogger
	ElectionTrigger      ElectionTrigger
	Storage              Storage
	//MessageFactory       *MessageFactory
}

type leanHelix struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (lh *leanHelix) RegisterOnCommitted(cb func(block Block)) {
	// TODO: implement
}

func (lh *leanHelix) Dispose() {

	// TODO: implement
}

func (lh *leanHelix) Start(blockHeight primitives.BlockHeight) {

	// TODO: create an infinite loop which can be stopped by context.Done()

	for {
		select {

		// case: some channel that fires when consensus completed successfully or with error
		case <-lh.ctx.Done():
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
		Logger:               config.Logger,
		BlockUtils:           config.BlockUtils,
	}
}
