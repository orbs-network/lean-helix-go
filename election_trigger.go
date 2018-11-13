package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTrigger interface {
	CreateElectionContextForView(parentContext context.Context, view primitives.View) context.Context
}
