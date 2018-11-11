package leanhelix

import "context"

type MessageReceiver interface {
	OnReceivePreprepare(ctx context.Context, ppm *PreprepareMessage)
	OnReceivePrepare(ctx context.Context, pm *PrepareMessage)
	OnReceiveCommit(ctx context.Context, cm *CommitMessage)
	OnReceiveViewChange(ctx context.Context, vcm *ViewChangeMessage)
	OnReceiveNewView(ctx context.Context, nvm *NewViewMessage)
}
