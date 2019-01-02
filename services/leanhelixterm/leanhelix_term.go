package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
)

type LeanHelixTerm struct {
	*ConsensusMessagesFilter
	termInCommittee *termincommittee.TermInCommittee
}

func NewLeanHelixTerm(ctx context.Context, config *interfaces.Config, onCommit interfaces.OnCommitCallback, prevBlock interfaces.Block) *LeanHelixTerm {
	randomSeed := uint64(12345)
	blockHeight := blockheight.GetBlockHeight(prevBlock) + 1
	messageFactory := messagesfactory.NewMessageFactory(config.KeyManager, config.Membership.MyMemberId(), randomSeed)

	// TODO: Implement the random seed
	committeeMembers := config.Membership.RequestOrderedCommittee(ctx, blockHeight, randomSeed)
	termInCommittee := termincommittee.NewTermInCommittee(ctx, config, messageFactory, committeeMembers, blockHeight, prevBlock, CommitsToProof(onCommit))

	return &LeanHelixTerm{
		ConsensusMessagesFilter: NewConsensusMessagesFilter(termInCommittee, config.KeyManager, randomSeed),
		termInCommittee:         termInCommittee,
	}
}
