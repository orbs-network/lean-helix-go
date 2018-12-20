package leanhelix

import (
	"context"
)

type TermInCommitteeMessagesHandler interface {
	HandleLeanHelixPrePrepare(ctx context.Context, ppm *PreprepareMessage)
	HandleLeanHelixPrepare(ctx context.Context, pm *PrepareMessage)
	HandleLeanHelixCommit(ctx context.Context, cm *CommitMessage)
	HandleLeanHelixViewChange(ctx context.Context, vcm *ViewChangeMessage)
	HandleLeanHelixNewView(ctx context.Context, nvm *NewViewMessage)
}
