package interfaces

import "github.com/orbs-network/lean-helix-go/spec/types/go/primitives"

var GenesisBlock Block = nil

type Block interface {
	Height() primitives.BlockHeight
}
