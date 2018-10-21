package leanhelix

import "context"

type MessageReceiver interface {
	OnReceivePreprepare(ctx context.Context, ppm PreprepareMessage) error
	OnReceivePrepare(ctx context.Context, pm PrepareMessage) error
	OnReceiveCommit(ctx context.Context, cm CommitMessage) error
	OnReceiveViewChange(ctx context.Context, vcm ViewChangeMessage) error
	OnReceiveNewView(ctx context.Context, nvm NewViewMessage) error
}
