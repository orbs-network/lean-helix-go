package leanhelix

import (
	"context"
	"fmt"
)

type LeanHelixTerm struct {
	termInCommittee *TermInCommittee
	onCommit        OnCommitCallback
}

func NewLeanHelixTerm(ctx context.Context, config *Config, onCommit OnCommitCallback, prevBlock Block) *LeanHelixTerm {
	result := &LeanHelixTerm{}

	blockHeight := GetBlockHeight(prevBlock) + 1

	// TODO: Implement the random seed
	committeeMembers := config.Membership.RequestOrderedCommittee(ctx, blockHeight, uint64(12345))
	termInCommittee := NewTermInCommittee(ctx, config, committeeMembers, result.onInCommitteeCommit, blockHeight, prevBlock)
	termInCommittee.StartTerm(ctx)

	result.termInCommittee = termInCommittee
	result.onCommit = onCommit

	return result
}

func (lht *LeanHelixTerm) onInCommitteeCommit(ctx context.Context, block Block, commitMessages []*CommitMessage) {
	lht.onCommit(ctx, block, GenerateLeanHelixBlockProof(commitMessages).Raw())
}

func (lht *LeanHelixTerm) HandleConsensusMessage(ctx context.Context, message ConsensusMessage) {
	if lht.termInCommittee == nil {
		return
	}

	switch message := message.(type) {
	case *PreprepareMessage:
		lht.termInCommittee.HandleLeanHelixPrePrepare(ctx, message)
	case *PrepareMessage:
		lht.termInCommittee.HandleLeanHelixPrepare(ctx, message)
	case *CommitMessage:
		lht.termInCommittee.HandleLeanHelixCommit(ctx, message)
	case *ViewChangeMessage:
		lht.termInCommittee.HandleLeanHelixViewChange(ctx, message)
	case *NewViewMessage:
		lht.termInCommittee.HandleLeanHelixNewView(ctx, message)
	default:
		panic(fmt.Sprintf("unknown message type: %T", message))
	}
}
