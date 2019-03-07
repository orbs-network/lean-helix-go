package leanhelixterm

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"math"
)

type LeanHelixTerm struct {
	*ConsensusMessagesFilter
	termInCommittee *termincommittee.TermInCommittee
}

func NewLeanHelixTerm(ctx context.Context, log logger.LHLogger, config *interfaces.Config, onCommit interfaces.OnCommitCallback, prevBlock interfaces.Block, prevBlockProofBytes []byte, canBeFirstLeader bool) *LeanHelixTerm {
	prevBlockProof := protocol.BlockProofReader(prevBlockProofBytes)
	randomSeed := randomseed.CalculateRandomSeed(prevBlockProof.RandomSeedSignature())
	blockHeight := blockheight.GetBlockHeight(prevBlock) + 1
	myMemberId := config.Membership.MyMemberId()
	messageFactory := messagesfactory.NewMessageFactory(config.InstanceId, config.KeyManager, myMemberId, randomSeed)

	committeeMembers, err := config.Membership.RequestOrderedCommittee(ctx, blockHeight, randomSeed)
	if err != nil {
		committeeMembers = nil // this will make sure isParticipating will be false (should be happening on system shutdown only)
		log.Info(L.LC(blockHeight, math.MaxUint64, myMemberId), "ERROR RECEIVING COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, members=%s", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, termincommittee.ToCommitteeMembersStr(committeeMembers))
	}

	isParticipating := isParticipatingInCommittee(myMemberId, committeeMembers)
	log.Info(L.LC(blockHeight, math.MaxUint64, myMemberId), "RECEIVED COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, members=%s, isParticipating=%t", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, termincommittee.ToCommitteeMembersStr(committeeMembers), isParticipating)
	if isParticipating {
		termInCommittee := termincommittee.NewTermInCommittee(ctx, log, config, messageFactory, committeeMembers, blockHeight, prevBlock, canBeFirstLeader, CommitsToProof(config.KeyManager, onCommit))
		return &LeanHelixTerm{
			ConsensusMessagesFilter: NewConsensusMessagesFilter(termInCommittee, config.KeyManager, randomSeed),
			termInCommittee:         termInCommittee,
		}
	} else {
		return &LeanHelixTerm{
			ConsensusMessagesFilter: NewConsensusMessagesFilter(nil, config.KeyManager, randomSeed),
			termInCommittee:         nil,
		}
	}
}

func (lht *LeanHelixTerm) Dispose() {
	if lht.termInCommittee != nil {
		lht.termInCommittee.Dispose()
		lht.termInCommittee = nil
	}
}

func isParticipatingInCommittee(myMemberId primitives.MemberId, committeeMembers []primitives.MemberId) bool {
	for _, committeeMember := range committeeMembers {
		if myMemberId.Equal(committeeMember) {
			return true
		}
	}
	return false
}

func printShortBlockProofBytes(b []byte) string {
	if len(b) < 6 {
		return ""
	}
	return fmt.Sprintf("%s..%s", hex.EncodeToString(b[:6]), hex.EncodeToString(b[len(b)-6:]))
}
