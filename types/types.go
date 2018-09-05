package types

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

type BlockUtils interface {
	CalculateBlockHash(block *Block) *BlockHash
}

type NetworkCommunication interface {
	SendToMembers(publicKeys []string, messageType string, message []byte)
}

// Sorting ViewCounter arrays
type ViewCounters []ViewCounter

func (arr ViewCounters) Len() int           { return len(arr) }
func (arr ViewCounters) Swap(i, j int)      { arr[i], arr[j] = arr[j], arr[i] }
func (arr ViewCounters) Less(i, j int) bool { return arr[i] < arr[j] }
