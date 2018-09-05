package inmemoryblockchain

import "github.com/orbs-network/lean-helix-go/types"

var GenesisBlock = &types.Block{
	Header: &types.BlockHeader{
		Height:    0,
		BlockHash: types.BlockHash("The Genesis Block"),
	},
	Body: []byte("The Genesis Block"),
}
