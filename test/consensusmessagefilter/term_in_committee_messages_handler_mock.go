package consensusmessagefilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
)

type termInCommitteeMessagesHandlerMock struct {
	HistoryPP []*leanhelix.PreprepareMessage
	HistoryP  []*leanhelix.PrepareMessage
	HistoryC  []*leanhelix.CommitMessage
	HistoryVC []*leanhelix.ViewChangeMessage
	HistoryNV []*leanhelix.NewViewMessage
}

func NewTermInCommitteeMessagesHandlerMock() *termInCommitteeMessagesHandlerMock {
	return &termInCommitteeMessagesHandlerMock{}
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixPrePrepare(ctx context.Context, ppm *leanhelix.PreprepareMessage) {
	t.HistoryPP = append(t.HistoryPP, ppm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixPrepare(ctx context.Context, pm *leanhelix.PrepareMessage) {
	t.HistoryP = append(t.HistoryP, pm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixCommit(ctx context.Context, cm *leanhelix.CommitMessage) {
	t.HistoryC = append(t.HistoryC, cm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixViewChange(ctx context.Context, vcm *leanhelix.ViewChangeMessage) {
	t.HistoryVC = append(t.HistoryVC, vcm)
}

func (t *termInCommitteeMessagesHandlerMock) HandleLeanHelixNewView(ctx context.Context, nvm *leanhelix.NewViewMessage) {
	t.HistoryNV = append(t.HistoryNV, nvm)
}
