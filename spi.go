package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

// This file contains SPI interfaces
// SPI is Service Programming Interface, these are the interfaces the consumer of this library
// must implement in order to use the library.

type Membership interface {
	MyMemberId() primitives.MemberId
	RequestOrderedCommittee(ctx context.Context, blockHeight primitives.BlockHeight, seed uint64, maxCommitteeSize uint32) []primitives.MemberId
}

type ConsensusRawMessage struct {
	Content []byte
	Block   Block
}

type Communication interface {
	SendConsensusMessage(ctx context.Context, targets []primitives.MemberId, message *ConsensusRawMessage)
}

type KeyManager interface {
	SignConsensusMessage(blockHeight primitives.BlockHeight, content []byte) primitives.Signature
	VerifyConsensusMessage(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) bool
	SignRandomSeed(blockHeight primitives.BlockHeight, content []byte) primitives.RandomSeedSignature
	VerifyRandomSeed(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) bool
	AggregateRandomSeed(blockHeight primitives.BlockHeight, randomSeedShares []*protocol.SenderSignature) primitives.RandomSeedSignature
}

type BlockUtils interface {
	RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock Block) (Block, primitives.BlockHash)
	ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, block Block, blockHash primitives.BlockHash, prevBlock Block) bool
	ValidateBlockCommitment(blockHeight primitives.BlockHeight, block Block, blockHash primitives.BlockHash) bool
}
