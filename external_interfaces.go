package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

type Block interface {
	GetHeight() primitives.BlockHeight
	GetBlockHash() primitives.Uint256
	//Body() []byte
}

type NetworkCommunication interface {
	SendToMembers(publicKeys []primitives.Ed25519PublicKey, messageType string, message []MessageTransporter)
	RequestOrderedCommittee(seed uint64) []primitives.Ed25519PublicKey
	IsMember(pk primitives.Ed25519PublicKey) bool

	Send(publicKeys []primitives.Ed25519PublicKey, message []byte) error
	SendWithBlock(publicKeys []primitives.Ed25519PublicKey, message []byte, block Block) error
}

type KeyManager interface {
	Sign(content []byte) []byte
	Verify(content []byte, sender *SenderSignature) bool
	MyPublicKey() primitives.Ed25519PublicKey
}

type BlockUtils interface {
	CalculateBlockHash(block Block) primitives.Uint256
	RequestNewBlock(blockHeight primitives.BlockHeight) Block
	ValidateBlock(block Block) bool
	RequestCommittee()
}
