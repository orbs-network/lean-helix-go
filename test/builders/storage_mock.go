package builders

import (
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
)

type MockStorage struct {
	mock.Mock
}

func (s *MockStorage) StorePreprepare(ppm leanhelix.PreprepareMessage) bool {
	ret := s.Called(ppm)
	return ret.Get(0).(bool)
}

func (s *MockStorage) StorePrepare(pp leanhelix.PrepareMessage) bool {
	//ret := s.Called(pp)
	//return ret.Get(0).(bool)
	panic("implement me")
}

func (s *MockStorage) StoreCommit(cm leanhelix.CommitMessage) bool {
	panic("implement me")
}

func (s *MockStorage) StoreViewChange(vcm leanhelix.ViewChangeMessage) bool {
	panic("implement me")
}

func (s *MockStorage) GetPrepareSendersPKs(term leanhelix.BlockHeight, view leanhelix.ViewCounter, blockHash leanhelix.BlockHash) []leanhelix.PublicKey {
	panic("implement me")
}

func (s *MockStorage) GetCommitSendersPKs(term leanhelix.BlockHeight, view leanhelix.ViewCounter, blockHash leanhelix.BlockHash) []leanhelix.PublicKey {
	panic("implement me")
}

func (s *MockStorage) GetViewChangeMessages(term leanhelix.BlockHeight, view leanhelix.ViewCounter, f int) []leanhelix.ViewChangeMessage {
	panic("implement me")
}

func (s *MockStorage) GetPreprepare(term leanhelix.BlockHeight, view leanhelix.ViewCounter) (leanhelix.PreprepareMessage, bool) {
	panic("implement me")
}

func (s *MockStorage) GetPrepares(term leanhelix.BlockHeight, view leanhelix.ViewCounter, blockHash leanhelix.BlockHash) ([]leanhelix.PrepareMessage, bool) {
	panic("implement me")
}

func (s *MockStorage) GetLatestPrepared(term leanhelix.BlockHeight, f int) (leanhelix.PreparedProof, bool) {
	panic("implement me")
}

func (s *MockStorage) ClearTermLogs(term leanhelix.BlockHeight) {
	panic("implement me")
}
