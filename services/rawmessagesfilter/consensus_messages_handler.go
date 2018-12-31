package rawmessagesfilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type ConsensusMessagesHandler interface {
	HandleConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage)
}
