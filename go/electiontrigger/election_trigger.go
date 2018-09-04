package electiontrigger

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

type ElectionTrigger interface {
	RegisterOnTrigger(view lh.ViewCounter, cb func(view lh.ViewCounter))
	UnregisterOnTrigger()
}
