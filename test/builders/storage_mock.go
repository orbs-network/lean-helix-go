package builders

import (
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
)

type mockStorage struct {
	mock.Mock
}

func NewMockStorage() *mockStorage {
	return &mockStorage{}
}

func (s *mockStorage) StorePreprepare(ppm leanhelix.PreprepareMessage) bool {
	ret := s.Called(ppm)
	return ret.Get(0).(bool)
}

func (s *mockStorage) StorePrepare(pp leanhelix.PrepareMessage) bool {
	//ret := s.Called(pp)
	//return ret.Get(0).(bool)
	panic("implement me")
}

func (s *mockStorage) StoreCommit(cm leanhelix.CommitMessage) bool {
	panic("implement me")
}

func (s *mockStorage) StoreViewChange(vcm leanhelix.ViewChangeMessage) bool {
	panic("implement me")
}

func (s *mockStorage) GetPrepareSendersPKs(term leanhelix.BlockHeight, view leanhelix.ViewCounter, blockHash leanhelix.BlockHash) []leanhelix.PublicKey {
	panic("implement me")
}

func (s *mockStorage) GetCommitSendersPKs(term leanhelix.BlockHeight, view leanhelix.ViewCounter, blockHash leanhelix.BlockHash) []leanhelix.PublicKey {
	panic("implement me")
}

func (s *mockStorage) GetViewChangeMessages(term leanhelix.BlockHeight, view leanhelix.ViewCounter, f int) []leanhelix.ViewChangeMessage {
	panic("implement me")
}

func (s *mockStorage) GetPreprepare(term leanhelix.BlockHeight, view leanhelix.ViewCounter) (leanhelix.PreprepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) GetPrepares(term leanhelix.BlockHeight, view leanhelix.ViewCounter, blockHash leanhelix.BlockHash) ([]leanhelix.PrepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) GetLatestPrepared(term leanhelix.BlockHeight, f int) (leanhelix.PreparedProof, bool) {
	panic("implement me")
}

func (s *mockStorage) ClearTermLogs(term leanhelix.BlockHeight) {
	panic("implement me")
}
