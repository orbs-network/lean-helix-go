package builders

import (
	"context"
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
)

type MockMessageReceiver struct {
	mock.Mock
}

func (rec *MockMessageReceiver) OnReceivePreprepare(ctx context.Context, ppm *leanhelix.PreprepareMessage) error {
	fmt.Println("OnReceivePreprepare")
	ret := rec.Called(ctx, ppm)
	return ret.Error(0)
}

func (rec *MockMessageReceiver) OnReceivePrepare(ctx context.Context, pm *leanhelix.PrepareMessage) error {
	fmt.Println("OnReceivePrepare")
	ret := rec.Called(ctx, pm)
	return ret.Error(0)
}

func (rec *MockMessageReceiver) OnReceiveCommit(ctx context.Context, cm *leanhelix.CommitMessage) error {
	ret := rec.Called(ctx, cm)
	return ret.Error(0)
}

func (rec *MockMessageReceiver) OnReceiveViewChange(ctx context.Context, vcm *leanhelix.ViewChangeMessage) error {
	ret := rec.Called(ctx, vcm)
	return ret.Error(0)
}

func (rec *MockMessageReceiver) OnReceiveNewView(ctx context.Context, nvm *leanhelix.NewViewMessage) error {
	ret := rec.Called(ctx, nvm)
	return ret.Error(0)
}

func NewMockMessageReceiver() *MockMessageReceiver {
	return &MockMessageReceiver{}
}
