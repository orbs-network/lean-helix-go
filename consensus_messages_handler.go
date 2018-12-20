package leanhelix

import (
	"context"
)

type ConsensusMessagesHandler interface {
	HandleConsensusMessage(ctx context.Context, message ConsensusMessage)
}
