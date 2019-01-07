package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type LeanHelixTerm struct {
	*ConsensusMessagesFilter
	termInCommittee *termincommittee.TermInCommittee
}

func NewLeanHelixTerm(ctx context.Context, config *interfaces.Config, onCommit interfaces.OnCommitCallback, prevBlock interfaces.Block, prevBlockProofBytes []byte) *LeanHelixTerm {
	prevBlockProof := protocol.BlockProofReader(prevBlockProofBytes)
	randomSeed := randomseed.CalculateRandomSeed(prevBlockProof.RandomSeedSignature())
	blockHeight := blockheight.GetBlockHeight(prevBlock) + 1
	messageFactory := messagesfactory.NewMessageFactory(config.KeyManager, config.Membership.MyMemberId(), randomSeed)

	committeeMembers := config.Membership.RequestOrderedCommittee(ctx, blockHeight, randomSeed)
	termInCommittee := termincommittee.NewTermInCommittee(ctx, config, messageFactory, committeeMembers, blockHeight, prevBlock, CommitsToProof(config.KeyManager, onCommit))

	return &LeanHelixTerm{
		ConsensusMessagesFilter: NewConsensusMessagesFilter(termInCommittee, config.KeyManager, randomSeed),
		termInCommittee:         termInCommittee,
	}
}
