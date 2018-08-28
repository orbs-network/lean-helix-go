package leanhelix

type BlockHeight uint64
type ViewCounter uint64
type BlockHash []byte
type PublicKey []byte

type BlockHeader struct {
	Height    BlockHeight
	BlockHash BlockHash
}

type Block struct {
	Header *BlockHeader
	Body   []byte
}

type Node struct {
	MyPublicKey PublicKey
}

type BlockUtils interface {
	CalculateBlockHash(block *Block) []byte
}

type NetworkCommunication interface {
	Nodes() []Node
	SendToMembers(publicKeys []string, messageType string, message []byte)
}
