package leanhelixterm

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/blockproof"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
)

type LeanHelixTerm struct {
	termInCommittee *termincommittee.TermInCommittee
	onCommit        interfaces.OnCommitCallback
}

func NewLeanHelixTerm(ctx context.Context, config *interfaces.Config, onCommit interfaces.OnCommitCallback, prevBlock interfaces.Block) *LeanHelixTerm {
	result := &LeanHelixTerm{}

	blockHeight := blockheight.GetBlockHeight(prevBlock) + 1

	messageFactory := messagesfactory.NewMessageFactory(config.KeyManager, config.Membership.MyMemberId())

	// TODO: Implement the random seed
	committeeMembers := config.Membership.RequestOrderedCommittee(ctx, blockHeight, uint64(12345))
	termInCommittee := termincommittee.NewTermInCommittee(ctx, config, messageFactory, committeeMembers, result.onInCommitteeCommit, blockHeight, prevBlock)
	termInCommittee.StartTerm(ctx)

	result.termInCommittee = termInCommittee
	result.onCommit = onCommit

	return result
}

func (lht *LeanHelixTerm) onInCommitteeCommit(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage) {
	lht.onCommit(ctx, block, blockproof.GenerateLeanHelixBlockProof(commitMessages).Raw())
}

func (lht *LeanHelixTerm) HandleConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	if lht.termInCommittee == nil {
		return
	}

	switch message := message.(type) {
	case *interfaces.PreprepareMessage:
		lht.termInCommittee.HandleLeanHelixPrePrepare(ctx, message)
	case *interfaces.PrepareMessage:
		lht.termInCommittee.HandleLeanHelixPrepare(ctx, message)
	case *interfaces.CommitMessage:
		lht.termInCommittee.HandleLeanHelixCommit(ctx, message)
	case *interfaces.ViewChangeMessage:
		lht.termInCommittee.HandleLeanHelixViewChange(ctx, message)
	case *interfaces.NewViewMessage:
		lht.termInCommittee.HandleLeanHelixNewView(ctx, message)
	default:
		panic(fmt.Sprintf("unknown message type: %T", message))
	}
}
