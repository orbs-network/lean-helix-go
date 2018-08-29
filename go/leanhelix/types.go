package leanhelix

type BlockHeight uint64
type ViewCounter uint64
type BlockHash string
type PublicKey string

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
	CalculateBlockHash(block *Block) *BlockHash
}

type NetworkCommunication interface {
	Nodes() []Node
	SendToMembers(publicKeys []string, messageType string, message []byte)
}

type PublicKeys []PublicKey

func (pks PublicKeys) Len() int {
	return len(pks)
}

func (pks PublicKeys) Less(i, j int) bool {
	return pks[i] < pks[j]
}

func (pks PublicKeys) Swap(i, j int) {
	pks[i], pks[j] = pks[j], pks[i]
}
