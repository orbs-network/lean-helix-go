package consensusmessagefilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
)

type termInCommitteeMessagesHandlerMock struct {
	historyPP []*leanhelix.PreprepareMessage
	historyP  []*leanhelix.PrepareMessage
	historyC  []*leanhelix.CommitMessage
	historyVC []*leanhelix.ViewChangeMessage
	historyNV []*leanhelix.NewViewMessage
}

func NewTermInCommitteeMessagesHandlerMock() *termInCommitteeMessagesHandlerMock {
	return &termInCommitteeMessagesHandlerMock{}
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixPrePrepare(ctx context.Context, ppm *leanhelix.PreprepareMessage) {
	t.historyPP = append(t.historyPP, ppm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixPrepare(ctx context.Context, pm *leanhelix.PrepareMessage) {
	t.historyP = append(t.historyP, pm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixCommit(ctx context.Context, cm *leanhelix.CommitMessage) {
	t.historyC = append(t.historyC, cm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixViewChange(ctx context.Context, vcm *leanhelix.ViewChangeMessage) {
	t.historyVC = append(t.historyVC, vcm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixNewView(ctx context.Context, nvm *leanhelix.NewViewMessage) {
	t.historyNV = append(t.historyNV, nvm)
}
