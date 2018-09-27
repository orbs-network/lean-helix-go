package builders

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
)

type MockBlockUtils struct {
	mock.Mock
	upcomingBlocks []lh.Block
}

func NewMockBlockUtils(upcomingBlocks []lh.Block) *MockBlockUtils {
	return &MockBlockUtils{
		upcomingBlocks: upcomingBlocks,
	}
}

func CalculateBlockHash(block lh.Block) lh.BlockHash {
	return lh.BlockHash(fmt.Sprintf("%s_%d_%s", block.Body(), block.Header().Term(), block.Header().BlockHash()))
}

func (*MockBlockUtils) CalculateBlockHash(block lh.Block) lh.BlockHash {
	return CalculateBlockHash(block)
}

func (mockBlockUtils *MockBlockUtils) ProvideNextBlock() {

}
func (mockBlockUtils *MockBlockUtils) ResolveAllValidations(b bool) {

}

func (mockBlockUtils *MockBlockUtils) RequestCommittee() {
	panic("implement me")
}

func (mockBlockUtils *MockBlockUtils) RequestNewBlock() {
	panic("implement me")
}

func (mockBlockUtils *MockBlockUtils) ValidateBlock() {
	panic("implement me")
}
