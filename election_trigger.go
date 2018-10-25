package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTrigger interface {
	RegisterOnTrigger(view primitives.View, cb func(ctx context.Context, view primitives.View))
	UnregisterOnTrigger()
}
