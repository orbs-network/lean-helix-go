package leanhelix

import (
	"context"
)

type TermMessagesHandler interface {
	HandleTermMessages(ctx context.Context, message ConsensusMessage)
}
