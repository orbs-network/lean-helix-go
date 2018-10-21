package builders

import (
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type mockStorage struct {
	mock.Mock
}

func (s *mockStorage) StorePreprepare(ppm *leanhelix.PreprepareMessage) bool {
	ret := s.Called(ppm)
	return ret.Get(0).(bool)
}

func (s *mockStorage) GetPreprepareMessage(blockHeight BlockHeight, view View) (*leanhelix.PreprepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) GetPreprepareBlock(blockHeight BlockHeight, view View) (leanhelix.Block, bool) {
	panic("implement me")
}

func (s *mockStorage) GetLatestPreprepare(blockHeight BlockHeight) (*leanhelix.PreprepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) StorePrepare(pp *leanhelix.PrepareMessage) bool {
	panic("implement me")
}

func (s *mockStorage) GetPrepareMessages(blockHeight BlockHeight, view View, blockHash Uint256) ([]*leanhelix.PrepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey {
	panic("implement me")
}

func (s *mockStorage) StoreCommit(cm *leanhelix.CommitMessage) bool {
	panic("implement me")
}

func (s *mockStorage) GetCommitMessages(blockHeight BlockHeight, view View, blockHash Uint256) ([]*leanhelix.CommitMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey {
	panic("implement me")
}

func (s *mockStorage) StoreViewChange(vcm *leanhelix.ViewChangeMessage) bool {
	panic("implement me")
}

func (s *mockStorage) GetViewChangeMessages(blockHeight BlockHeight, view View) []*leanhelix.ViewChangeMessage {
	panic("implement me")
}

func (s *mockStorage) ClearBlockHeightLogs(blockHeight BlockHeight) {
	panic("implement me")
}

func NewMockStorage() *mockStorage {
	return &mockStorage{}
}
