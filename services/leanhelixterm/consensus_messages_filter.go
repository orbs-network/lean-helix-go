package leanhelixterm

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type ConsensusMessagesFilter struct {
	handler    TermMessagesHandler
	keyManager interfaces.KeyManager
	randomSeed uint64
}

func NewConsensusMessagesFilter(handler TermMessagesHandler, keyManager interfaces.KeyManager, randomSeed uint64) *ConsensusMessagesFilter {
	return &ConsensusMessagesFilter{handler, keyManager, randomSeed}
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
		senderSignature := (&protocol.SenderSignatureBuilder{
			MemberId:  message.Content().Sender().MemberId(),
			Signature: primitives.Signature(message.Content().Share()),
		}).Build()

		randomSeedBytes := randomseed.RandomSeedToBytes(mp.randomSeed)
		if !mp.keyManager.VerifyRandomSeed(message.BlockHeight(), randomSeedBytes, senderSignature) {
			//fmt.Println("Filter VerifyRandomSeed Failed")
			return
		}
		mp.handler.HandleCommit(ctx, message)

	case *interfaces.ViewChangeMessage:
		mp.handler.HandleViewChange(ctx, message)

	case *interfaces.NewViewMessage:
		mp.handler.HandleNewView(ctx, message)

	default:
		panic(fmt.Sprintf("unknown message type: %T", message))
	}
}
