package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

type Block interface {
	Height() primitives.BlockHeight
	BlockHash() primitives.Uint256
	PrevBlockHash() primitives.Uint256
	//Body() []byte
}

type ConsensusMessage interface {
	Content() []byte
	Block() Block
}

type NetworkCommunication interface {
	RequestOrderedCommittee(seed uint64) []primitives.Ed25519PublicKey
	IsMember(pk primitives.Ed25519PublicKey) bool

	Send(targets []primitives.Ed25519PublicKey, message ConsensusMessage) error
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
