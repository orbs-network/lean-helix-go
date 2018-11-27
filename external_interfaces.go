package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/primitives"
)

type Config struct {
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	ElectionTrigger      ElectionTrigger
	Storage              Storage
}

// Interfaces that must be implemented by the external service using this library

// A block instance for which library tries to reach consensus
type Block interface {
	Height() primitives.BlockHeight
}
