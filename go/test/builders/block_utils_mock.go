package builders

import (
	"github.com/orbs-network/lean-helix-go/go/block"
	"github.com/stretchr/testify/mock"
)

type MockBlockUtils struct {
	mock.Mock
}

func (*MockBlockUtils) CalculateBlockHash(block *block.Block) []byte {
	return []byte("0123456789ABCDEF")
}
