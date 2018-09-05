package leanhelix

import "github.com/orbs-network/lean-helix-go/types"

type ElectionTrigger interface {
	RegisterOnTrigger(view types.ViewCounter, cb func(view types.ViewCounter))
	UnregisterOnTrigger()
}
