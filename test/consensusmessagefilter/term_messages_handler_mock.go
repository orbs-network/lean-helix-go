package consensusmessagefilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
)

type termMessagesHandlerMock struct {
	historyPP []*leanhelix.PreprepareMessage
	historyP  []*leanhelix.PrepareMessage
	historyC  []*leanhelix.CommitMessage
	historyVC []*leanhelix.ViewChangeMessage
	historyNV []*leanhelix.NewViewMessage
}

func NewTermMessagesHandlerMock() *termMessagesHandlerMock {
	return &termMessagesHandlerMock{}
}

func (t *termMessagesHandlerMock) OnReceivePreprepare(ctx context.Context, ppm *leanhelix.PreprepareMessage) {
	t.historyPP = append(t.historyPP, ppm)
}

func (t *termMessagesHandlerMock) OnReceivePrepare(ctx context.Context, pm *leanhelix.PrepareMessage) {
	t.historyP = append(t.historyP, pm)
}

func (t *termMessagesHandlerMock) OnReceiveCommit(ctx context.Context, cm *leanhelix.CommitMessage) {
	t.historyC = append(t.historyC, cm)
}

func (t *termMessagesHandlerMock) OnReceiveViewChange(ctx context.Context, vcm *leanhelix.ViewChangeMessage) {
	t.historyVC = append(t.historyVC, vcm)
}

func (t *termMessagesHandlerMock) OnReceiveNewView(ctx context.Context, nvm *leanhelix.NewViewMessage) {
	t.historyNV = append(t.historyNV, nvm)
}
