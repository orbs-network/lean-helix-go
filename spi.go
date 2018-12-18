package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

// This file contains SPI interfaces
// SPI is Service Programming Interface, these are the interfaces the consumer of this library
// must implement in order to use the library.

type LeanHelixSPI struct {
	Utils BlockUtils
	Comm  Communication
	Mgr   KeyManager
}

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

type RandomSeedShare struct {
	signature primitives.Signature
	memberId  primitives.MemberId
}

type KeyManager interface {
	SignConsensusMessage(content []byte) []byte
	VerifyConsensusMessage(content []byte, signature primitives.Signature, memberId primitives.MemberId) bool
	SignRandomSeed(blockHeight primitives.BlockHeight, content []byte) []byte
	VerifyRandomSeed(blockHeight primitives.BlockHeight, content []byte, signature primitives.Signature, memberId primitives.MemberId) bool
	AggregateRandomSeed(blockHeight primitives.BlockHeight, randomSeedShares []*RandomSeedShare) primitives.Signature
}

type BlockUtils interface {
	CalculateBlockHash(block Block) primitives.BlockHash
	RequestNewBlock(ctx context.Context, prevBlock Block) Block
	ValidateBlock(block Block) bool
}
