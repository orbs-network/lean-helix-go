package builders

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

type MockBlockUtils struct {
	mock.Mock
}

func NewMockBlockUtils() *MockBlockUtils {
	return &MockBlockUtils{}
}

func CalculateBlockHash(block *lh.Block) lh.BlockHash {
	return lh.BlockHash(fmt.Sprintf("%s_%d_%s", block.Body, block.Header.Height, block.Header.BlockHash))
}

func (*MockBlockUtils) CalculateBlockHash(block *lh.Block) lh.BlockHash {
	return CalculateBlockHash(block)
}
