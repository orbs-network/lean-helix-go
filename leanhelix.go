package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
)

// PBFT.ts
type LeanHelix interface {
	RegisterOnCommitted(cb func(block Block))
	Dispose()
	Start(height BlockHeight)
	IsLeader() bool
}

// TODO looks identical to Config, why is this needed?
type TermConfig struct {
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	Logger               log.BasicLogger
	ElectionTrigger      ElectionTrigger
	Storage              Storage
}

type leanHelix struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	log       log.BasicLogger
}

func NewLeanHelix(config *Config) LeanHelix {

	ctx, ctxCancel := context.WithCancel(config.ctx)

	return &leanHelix{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		log:       config.Logger.For(log.Service("leanhelix")),
	}
}

func (lh *leanHelix) RegisterOnCommitted(cb func(block Block)) {
	// TODO: implement
}

func (lh *leanHelix) Dispose() {

	// TODO: implement
}

func (lh *leanHelix) Start(height BlockHeight) {

	// TODO: create an infinite loop which can be stopped by context.Done()

	for {
		select {

		// case: some channel that fires when consensus completed successfully or with error
		case <-lh.ctx.Done():
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
	lh.log.Info("GracefulShutdown() called")
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
