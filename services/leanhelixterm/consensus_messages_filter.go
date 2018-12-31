package leanhelixterm

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type ConsensusMessagesFilter struct {
	handler TermMessagesHandler
}

func NewConsensusMessagesFilter(handler TermMessagesHandler) *ConsensusMessagesFilter {
	return &ConsensusMessagesFilter{handler}
}

func (mp *ConsensusMessagesFilter) HandleConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	if mp.handler == nil {
		return
	}

	switch message := message.(type) {
	case *interfaces.PreprepareMessage:
		mp.handler.HandlePrePrepare(ctx, message)
	case *interfaces.PrepareMessage:
		mp.handler.HandlePrepare(ctx, message)
	case *interfaces.CommitMessage:
		mp.handler.HandleCommit(ctx, message)
	case *interfaces.ViewChangeMessage:
		mp.handler.HandleViewChange(ctx, message)
	case *interfaces.NewViewMessage:
		mp.handler.HandleNewView(ctx, message)
	default:
		panic(fmt.Sprintf("unknown message type: %T", message))
	}
}
