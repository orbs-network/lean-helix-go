package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

// This file contains SPI interfaces
// SPI is Service Programming Interface, these are the interfaces the consumer of this library
// must implement in order to use the library.

type LeanHelixSPI struct {
	Utils BlockUtils
	Comm  NetworkCommunication
	Mgr   KeyManager
}

// Communication layer for sending & receiving messages, and requesting committee and checking committee membership
type NetworkCommunication interface {
	RequestOrderedCommittee(ctx context.Context, blockHeight primitives.BlockHeight, seed uint64, maxCommitteeSize uint32) []primitives.MemberId
	SendMessage(ctx context.Context, targets []primitives.MemberId, message ConsensusRawMessage)
}

type KeyManager interface {
	Sign(content []byte) []byte
	Verify(content []byte, sender *protocol.SenderSignature) bool
	MyMemberId() primitives.MemberId
}

type BlockUtils interface {
	CalculateBlockHash(block Block) primitives.BlockHash
	RequestNewBlock(ctx context.Context, prevBlock Block) Block
	ValidateBlock(block Block) bool
}
