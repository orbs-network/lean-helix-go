package leanhelix

import "github.com/orbs-network/lean-helix-go/instrumentation/log"

// PBFT.ts
type LeanHelix interface {
	RegisterOnCommitted(cb func(block Block))
	Dispose()
	Start(height BlockHeight)
	IsLeader() bool
}

// TODO looks indentical to Config, why is this needed?
type TermConfig struct {
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	Logger               log.BasicLogger
	ElectionTrigger      ElectionTrigger
	Storage              Storage
}

type leanHelix struct {
}

func NewLeanHelix(config *Config) LeanHelix {
	return &leanHelix{}
}

func (lh *leanHelix) RegisterOnCommitted(cb func(block Block)) {
	// TODO: implement
}

func (lh *leanHelix) Dispose() {
	// TODO: implement
}

func (lh *leanHelix) Start(height BlockHeight) {
	// TODO: implement
}

func (lh *leanHelix) IsLeader() bool {
	// TODO: implement
	return false
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
