package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

type ElectionTrigger interface {
	RegisterOnTrigger(view primitives.View, cb func(view primitives.View))
	UnregisterOnTrigger()
}
