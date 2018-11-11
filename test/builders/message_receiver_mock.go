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

func (rec *MockMessageReceiver) OnReceivePreprepare(ctx context.Context, ppm *leanhelix.PreprepareMessage) {
	fmt.Println("OnReceivePreprepare")
	rec.Called(ctx, ppm)
}

func (rec *MockMessageReceiver) OnReceivePrepare(ctx context.Context, pm *leanhelix.PrepareMessage) {
	fmt.Println("OnReceivePrepare")
	rec.Called(ctx, pm)
}

func (rec *MockMessageReceiver) OnReceiveCommit(ctx context.Context, cm *leanhelix.CommitMessage) {
	rec.Called(ctx, cm)
}

func (rec *MockMessageReceiver) OnReceiveViewChange(ctx context.Context, vcm *leanhelix.ViewChangeMessage) {
	rec.Called(ctx, vcm)
}

func (rec *MockMessageReceiver) OnReceiveNewView(ctx context.Context, nvm *leanhelix.NewViewMessage) {
	rec.Called(ctx, nvm)
}

func NewMockMessageReceiver() *MockMessageReceiver {
	return &MockMessageReceiver{}
}
