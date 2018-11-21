package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

// This file contains SPI interfaces
// SPI is Service Programming Interface, these are the interfaces the consumer of this library
// must implement in order to use the library.

type LeanHelixSPI struct {
	Utils BlockUtils
	Comm  NetworkCommunication
	Mgr   KeyManager
}

type MessageHandler func(ctx context.Context, message ConsensusRawMessage)

// Communication layer for sending & receiving messages, and requesting committee and checking committee membership
type NetworkCommunication interface {
	RequestOrderedCommittee(seed uint64) []primitives.Ed25519PublicKey
	IsMember(pk primitives.Ed25519PublicKey) bool
	RegisterOnMessage(onReceivedMessage MessageHandler) int
	UnregisterOnMessage(subscriptionToken int)
	SendMessage(ctx context.Context, targets []primitives.Ed25519PublicKey, message ConsensusRawMessage)
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
