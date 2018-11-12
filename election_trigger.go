package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type ElectionTrigger interface {
	CreateElectionContext(parentContext context.Context, view primitives.View) context.Context
}
