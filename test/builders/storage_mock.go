package builders

import (
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type mockStorage struct {
	mock.Mock
}

func (s *mockStorage) GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey {
	panic("implement me")
}

func (s *mockStorage) GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey {
	panic("implement me")
}

func (s *mockStorage) GetPrepares(blockHeight BlockHeight, view View, blockHash Uint256) ([]leanhelix.PrepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) GetLatestPreprepare(blockHeight BlockHeight) (leanhelix.PreprepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) StorePreprepare(ppm leanhelix.PreprepareMessage) bool {
	ret := s.Called(ppm)
	return ret.Get(0).(bool)
}

func (s *mockStorage) StorePrepare(pp leanhelix.PrepareMessage) bool {
	panic("implement me")
}

func (s *mockStorage) StoreCommit(cm leanhelix.CommitMessage) bool {
	panic("implement me")
}

func (s *mockStorage) StoreViewChange(vcm leanhelix.ViewChangeMessage) bool {
	panic("implement me")
}

func (s *mockStorage) GetViewChangeMessages(blockHeight BlockHeight, view View, f int) []leanhelix.ViewChangeMessage {
	panic("implement me")
}

func (s *mockStorage) GetPreprepare(blockHeight BlockHeight, view View) (leanhelix.PreprepareMessage, bool) {
	panic("implement me")
}

func (s *mockStorage) GetLatestPrepared(blockHeight BlockHeight, f int) (leanhelix.PreparedProof, bool) {
	panic("implement me")
}

func (s *mockStorage) ClearTermLogs(blockHeight BlockHeight) {
	panic("implement me")
}

func NewMockStorage() *mockStorage {
	return &mockStorage{}
}
