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
	return ret.Bool(0)
}

func (s *mockStorage) GetPreprepareMessage(blockHeight BlockHeight, view View) (*leanhelix.PreprepareMessage, bool) {
	ret := s.Called(blockHeight, view)
	return ret.Get(0).(*leanhelix.PreprepareMessage), ret.Bool(1)
}

func (s *mockStorage) GetPreprepareBlock(blockHeight BlockHeight, view View) (leanhelix.Block, bool) {
	ret := s.Called(blockHeight, view)
	return ret.Get(0).(leanhelix.Block), ret.Bool(1)
}

func (s *mockStorage) GetLatestPreprepare(blockHeight BlockHeight) (*leanhelix.PreprepareMessage, bool) {
	ret := s.Called(blockHeight)
	return ret.Get(0).(*leanhelix.PreprepareMessage), ret.Bool(1)
}

func (s *mockStorage) StorePrepare(pp *leanhelix.PrepareMessage) bool {
	ret := s.Called(pp)
	return ret.Bool(0)
}

func (s *mockStorage) GetPrepareMessages(blockHeight BlockHeight, view View, blockHash Uint256) ([]*leanhelix.PrepareMessage, bool) {
	ret := s.Called(blockHeight, view)
	return ret.Get(0).([]*leanhelix.PrepareMessage), ret.Bool(1)

}

func (s *mockStorage) GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey {
	ret := s.Called(blockHeight, view, blockHash)
	return ret.Get(0).([]Ed25519PublicKey)
}

func (s *mockStorage) StoreCommit(cm *leanhelix.CommitMessage) bool {
	ret := s.Called(cm)
	return ret.Bool(0)
}

func (s *mockStorage) GetCommitMessages(blockHeight BlockHeight, view View, blockHash Uint256) ([]*leanhelix.CommitMessage, bool) {
	ret := s.Called(blockHeight, view, blockHash)
	return ret.Get(0).([]*leanhelix.CommitMessage), ret.Bool(1)
}

func (s *mockStorage) GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey {
	ret := s.Called(blockHeight, view, blockHash)
	return ret.Get(0).([]Ed25519PublicKey)
}

func (s *mockStorage) StoreViewChange(vcm *leanhelix.ViewChangeMessage) bool {
	ret := s.Called(vcm)
	return ret.Bool(0)
}

func (s *mockStorage) GetViewChangeMessages(blockHeight BlockHeight, view View) []*leanhelix.ViewChangeMessage {
	ret := s.Called(blockHeight, view)
	return ret.Get(0).([]*leanhelix.ViewChangeMessage)
}

func (s *mockStorage) ClearBlockHeightLogs(blockHeight BlockHeight) {
	s.Called(blockHeight)
}

func NewMockStorage() *mockStorage {
	return &mockStorage{}
}
