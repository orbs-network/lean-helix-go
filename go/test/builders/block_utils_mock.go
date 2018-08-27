package builders

import (
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go/go/block"
	"github.com/orbs-network/lean-helix-go/go/leanhelix"
)

type MockBlockUtils struct {
	mock.Mock
}

func NewMockBlockUtils() leanhelix.BlockUtils {
	return &MockBlockUtils{}
}

func (*MockBlockUtils) CalculateBlockHash(block *block.Block) []byte {
	return []byte("0123456789ABCDEF")
}
