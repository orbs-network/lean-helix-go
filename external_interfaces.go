package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	"github.com/orbs-network/lean-helix-go/primitives"
)

// first call - create an instance of Lean Helix library
func NewLeanHelix(config *Config) LeanHelix {

	return &leanHelix{
		log: config.Logger.For(log.Service("leanhelix")),
	}
}

type LeanHelix interface {
	Start(blockHeight primitives.BlockHeight)
	RegisterOnCommitted(cb func(block Block))
	Dispose()
	IsLeader() bool
}

type Config struct {
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	Logger               log.BasicLogger
	ElectionTrigger      ElectionTrigger
	Storage              Storage
}

// Interfaces that must be implemented by the external service using this library

// A block instance for which library tries to reach consensus
type Block interface {
	Height() primitives.BlockHeight
	BlockHash() primitives.Uint256
}

// Communication layer for sending & receiving messages, and requesting committee and checking committee membership
type NetworkCommunication interface {
	RequestOrderedCommittee(seed uint64) []primitives.Ed25519PublicKey
	IsMember(pk primitives.Ed25519PublicKey) bool
	// Register a callback to be called by the external service when a message consensus is received
	RegisterOnMessage(onReceivedMessage func(ctx context.Context, message ConsensusRawMessage)) int
	SendMessage(ctx context.Context, targets []primitives.Ed25519PublicKey, message ConsensusRawMessage) error
}

type KeyManager interface {
	Sign(content []byte) []byte
	Verify(content []byte, sender *SenderSignature) bool
	MyPublicKey() primitives.Ed25519PublicKey
}

type BlockUtils interface {
	CalculateBlockHash(block Block) primitives.Uint256
	RequestNewBlock(ctx context.Context, blockHeight primitives.BlockHeight) Block
	ValidateBlock(block Block) bool
}
