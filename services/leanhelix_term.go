package services

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockproof"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
)

func GetBlockHeight(prevBlock interfaces.Block) primitives.BlockHeight {
	if prevBlock == interfaces.GenesisBlock {
		return 0
	} else {
		return prevBlock.Height()
	}
}

func CalcQuorumSize(committeeMembersCount int) int {
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}

type LeanHelixTerm struct {
	termInCommittee *TermInCommittee
	onCommit        interfaces.OnCommitCallback
}

func NewLeanHelixTerm(ctx context.Context, config *interfaces.Config, onCommit interfaces.OnCommitCallback, prevBlock interfaces.Block) *LeanHelixTerm {
	result := &LeanHelixTerm{}

	blockHeight := GetBlockHeight(prevBlock) + 1

	messageFactory := messagesfactory.NewMessageFactory(config.KeyManager, config.Membership.MyMemberId())

	// TODO: Implement the random seed
	committeeMembers := config.Membership.RequestOrderedCommittee(ctx, blockHeight, uint64(12345))
	termInCommittee := NewTermInCommittee(ctx, config, messageFactory, committeeMembers, result.onInCommitteeCommit, blockHeight, prevBlock)
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
