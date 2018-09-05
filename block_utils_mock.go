package leanhelix

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go/types"
)

type MockBlockUtils struct {
	mock.Mock
}

func NewMockBlockUtils() *MockBlockUtils {
	return &MockBlockUtils{}
}

func CalculateBlockHash(block *types.Block) types.BlockHash {
	return types.BlockHash(fmt.Sprintf("%s_%d_%s", block.Body, block.Header.Height, block.Header.BlockHash))
}

func (*MockBlockUtils) CalculateBlockHash(block *types.Block) types.BlockHash {
	return CalculateBlockHash(block)
}
