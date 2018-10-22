package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

// Implemented by external service
type Block interface {
	Height() primitives.BlockHeight
	BlockHash() primitives.Uint256
	PrevBlockHash() primitives.Uint256
}

type ConsensusRawMessage interface {
	MessageType() MessageType
	Content() []byte
	Block() Block
	ToConsensusMessage() ConsensusMessage
}

// Implemented by external service
type NetworkCommunication interface {
	RequestOrderedCommittee(seed uint64) []primitives.Ed25519PublicKey
	IsMember(pk primitives.Ed25519PublicKey) bool
	RegisterOnMessage(func(ctx context.Context, message ConsensusRawMessage)) int
	SendMessage(ctx context.Context, targets []primitives.Ed25519PublicKey, message ConsensusRawMessage) error
}

// Implemented by external service
type KeyManager interface {
	Sign(content []byte) []byte
	Verify(content []byte, sender *SenderSignature) bool
	MyPublicKey() primitives.Ed25519PublicKey
}

// Implemented by external service
type BlockUtils interface {
	CalculateBlockHash(block Block) primitives.Uint256
	RequestNewBlock(ctx context.Context, blockHeight primitives.BlockHeight) Block
	ValidateBlock(block Block) bool
	RequestCommittee()
}
