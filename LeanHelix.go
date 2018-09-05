package leanhelix

import "github.com/orbs-network/lean-helix-go/types"

type LeanHelix struct {
}

func NewLeanHelix(config *Config) *LeanHelix {
	return &LeanHelix{}
}

func (pbft *LeanHelix) RegisterOnCommitted(cb func(block *types.Block)) {
	// TODO: implement
}

func (pbft *LeanHelix) Dispose() {
	// TODO: implement
}

func (pbft *LeanHelix) Start(height types.BlockHeight) {
	// TODO: implement
}

func (pbft *LeanHelix) IsLeader() bool {
	// TODO: implement
	return false
}
