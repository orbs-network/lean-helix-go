package leanhelix

type Block interface {
	GetHeight() BlockHeight
	GetBlockHash() BlockHash
	//Body() []byte
}

type NetworkCommunication interface {
	SendToMembers(publicKeys []PublicKey, messageType string, message []MessageTransporter)
	RequestOrderedCommittee(seed uint64) []PublicKey
	IsMember(pk PublicKey) bool

	Send(publicKeys []PublicKey, message []byte) error
	SendWithBlock(publicKeys []PublicKey, message []byte, block Block) error
}

type KeyManager interface {
	Sign(content []byte) []byte
	Verify(content []byte, sender SenderSignature) bool
	MyPublicKey() PublicKey
}

type BlockUtils interface {
	CalculateBlockHash(block Block) BlockHash
	RequestNewBlock(blockHeight BlockHeight) Block
	ValidateBlock(block Block) bool
	RequestCommittee()
}
