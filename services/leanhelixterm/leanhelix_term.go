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
	messageFactory := messagesfactory.NewMessageFactory(config.InstanceId, config.KeyManager, config.Membership.MyMemberId(), randomSeed)

	committeeMembers := config.Membership.RequestOrderedCommittee(ctx, blockHeight, randomSeed)
	log.Info(L.LC(blockHeight, math.MaxUint64, config.Membership.MyMemberId()), "RECEIVED COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, members=%s", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, termincommittee.ToCommitteeMembersStr(committeeMembers))
	termInCommittee := termincommittee.NewTermInCommittee(ctx, log, config, messageFactory, committeeMembers, blockHeight, prevBlock, canBeFirstLeader, CommitsToProof(config.KeyManager, onCommit))

	return &LeanHelixTerm{
		ConsensusMessagesFilter: NewConsensusMessagesFilter(termInCommittee, config.KeyManager, randomSeed),
		termInCommittee:         termInCommittee,
	}
}

func printShortBlockProofBytes(b []byte) string {
	if len(b) < 6 {
		return ""
	}
	return fmt.Sprintf("%s..%s", hex.EncodeToString(b[:6]), hex.EncodeToString(b[len(b)-6:]))
}
