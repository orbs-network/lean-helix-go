package block

type Block struct {
	Header *BlockHeader
	Body   string
}

type BlockHeader struct {
	Height        uint64
	PrevBlockHash []byte
}
