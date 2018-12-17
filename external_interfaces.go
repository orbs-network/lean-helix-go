package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

// The algorithm cannot function with less committee members
// because it cannot calculate the f number (where committee members are 3f+1)
// The only reason to set this manually in config below this limit is for internal tests
const LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS = 4

type Config struct {
	Communication                   Communication
	Membership                      Membership
	BlockUtils                      BlockUtils
	KeyManager                      KeyManager
	ElectionTrigger                 ElectionTrigger
	Storage                         Storage
	Logger                          Logger
	OverrideMinimumCommitteeMembers int
}

// Interfaces that must be implemented by the external service using this library

// A block instance for which library tries to reach consensus
type Block interface {
	Height() primitives.BlockHeight
}
