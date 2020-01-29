// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelixterm

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/scribe/log"
	"math"
)

type LeanHelixTerm struct {
	*ConsensusMessagesFilter
	termInCommittee *termincommittee.TermInCommittee
}

func NewLeanHelixTerm(ctx context.Context, logger logger.LHLogger, config *interfaces.Config, state *state.State, electionTrigger interfaces.ElectionScheduler, onCommit interfaces.OnCommitCallback, prevBlock interfaces.Block, prevBlockProofBytes []byte, canBeFirstLeader bool) *LeanHelixTerm {
	prevBlockProof := protocol.BlockProofReader(prevBlockProofBytes)
	randomSeed := randomseed.CalculateRandomSeed(prevBlockProof.RandomSeedSignature())
	blockHeight := blockheight.GetBlockHeight(prevBlock) + 1
	myMemberId := config.Membership.MyMemberId()
	messageFactory := messagesfactory.NewMessageFactory(config.InstanceId, config.KeyManager, myMemberId, randomSeed)

	committeeMembers, err := requestOrderedCommittee(state, blockHeight, randomSeed, config)
	if err != nil {
		logger.Info("OUT OF COMMITTEE WITH ERROR RECEIVING COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, error=%s", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, err)
		return termNotInCommittee(randomSeed, config)
	}

	isParticipating := isParticipatingInCommittee(myMemberId, committeeMembers)
	logger.Debug("RECEIVED COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, members=%s, isParticipating=%t", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, termincommittee.ToCommitteeMembersStr(committeeMembers), isParticipating)
	logger.ConsensusTrace("got committee for the current consensus round", nil, log.StringableSlice("committee", committeeMembers))

	if !isParticipating {
		logger.Info("OUT OF COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, members=%s, isParticipating=%t", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, termincommittee.ToCommitteeMembersStr(committeeMembers), isParticipating)
		return termNotInCommittee(randomSeed, config)
	}

	termInCommittee := termincommittee.NewTermInCommittee(logger, config, state, messageFactory, electionTrigger, committeeMembers, prevBlock, canBeFirstLeader, CommitsToProof(logger, config.KeyManager, onCommit))
	return &LeanHelixTerm{
		ConsensusMessagesFilter: NewConsensusMessagesFilter(termInCommittee, config.KeyManager, randomSeed),
		termInCommittee:         termInCommittee,
	}
}

func requestOrderedCommittee(s *state.State, blockHeight primitives.BlockHeight, randomSeed uint64, config *interfaces.Config) ([]primitives.MemberId, error) {
	const maxView = primitives.View(math.MaxUint64)
	ctx, err := s.Contexts.For(state.NewHeightView(blockHeight, maxView)) // term-level context
	if err != nil {
		return nil, err
	}
	committeeMembers, err := config.Membership.RequestOrderedCommittee(ctx, blockHeight, randomSeed)
	if err != nil {
		return nil, err
	}
	return committeeMembers, nil
}

func termNotInCommittee(randomSeed uint64, config *interfaces.Config) *LeanHelixTerm {
	return &LeanHelixTerm{
		ConsensusMessagesFilter: NewConsensusMessagesFilter(nil, config.KeyManager, randomSeed),
		termInCommittee:         nil,
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
