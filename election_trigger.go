package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTrigger interface {
	RegisterOnElection(view primitives.View, cb func(ctx context.Context, view primitives.View))
	ElectionChannel() chan func(ctx context.Context)
}
