package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

type ElectionTrigger interface {
	RegisterOnElection(view primitives.View, cb func(view primitives.View))
	ElectionChannel() chan func()
}
