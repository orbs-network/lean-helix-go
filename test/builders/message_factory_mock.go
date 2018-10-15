package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type MockMessageFactory struct {
	CalculateBlockHash func(block leanhelix.Block) primitives.Uint256
	KeyManager         leanhelix.KeyManager
}

func NewMockMessageFactory(calcHash func(block leanhelix.Block) primitives.Uint256, keyManager leanhelix.KeyManager) *MockMessageFactory {
	return &MockMessageFactory{
		CalculateBlockHash: calcHash,
		KeyManager:         keyManager,
	}
}

func (f *MockMessageFactory) CreatePreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) leanhelix.PreprepareMessage {
	panic("implement me")
}

func (f *MockMessageFactory) CreatePrepareMessage(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.Uint256) leanhelix.PrepareMessage {
	panic("implement me")
}

func (f *MockMessageFactory) CreateCommitMessage(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.Uint256) leanhelix.CommitMessage {
	panic("implement me")
}

func (f *MockMessageFactory) CreateViewChangeMessage(blockHeight primitives.BlockHeight, view primitives.View, preparedMessages *leanhelix.PreparedMessages) leanhelix.ViewChangeMessage {
	panic("implement me")
}

func (f *MockMessageFactory) CreateNewViewMessage(blockHeight primitives.BlockHeight, view primitives.View, ppm leanhelix.PreprepareMessage, confirmations []leanhelix.ViewChangeConfirmation) leanhelix.NewViewMessage {
	panic("implement me")
}
