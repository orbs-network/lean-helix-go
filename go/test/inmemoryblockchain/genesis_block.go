package inmemoryblockchain

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

var GenesisBlock = &lh.Block{
	Header: &lh.BlockHeader{
		Height:    0,
		BlockHash: lh.BlockHash("The Genesis Block"),
	},
	Body: []byte("The Genesis Block"),
}
