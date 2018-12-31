package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type TermMessagesHandler interface {
	HandlePrePrepare(ctx context.Context, ppm *interfaces.PreprepareMessage)
	HandlePrepare(ctx context.Context, pm *interfaces.PrepareMessage)
	HandleViewChange(ctx context.Context, vcm *interfaces.ViewChangeMessage)
	HandleCommit(ctx context.Context, cm *interfaces.CommitMessage)
	HandleNewView(ctx context.Context, nvm *interfaces.NewViewMessage)
}
