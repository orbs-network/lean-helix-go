package electiontrigger

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

type ElectionTrigger interface {
	Start(cb func())
	Stop()
}

type NewElectionTrigger = func(view lh.ViewCounter) ElectionTrigger
