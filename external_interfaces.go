package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type LeanHelix interface {
	Start(ctx context.Context, blockHeight primitives.BlockHeight)
	RegisterOnCommitted(cb func(block Block))
	ValidateBlockConsensus(block Block, blockProof *BlockProof, prevBlockProof *BlockProof)
}

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
