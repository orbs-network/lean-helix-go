package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type ElectionTrigger interface {
	RegisterOnElection(blockHeight primitives.BlockHeight, view primitives.View, cb func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View))
	ElectionChannel() chan func(ctx context.Context)
}
