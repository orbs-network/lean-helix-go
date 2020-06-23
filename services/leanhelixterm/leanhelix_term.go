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
	"github.com/orbs-network/lean-helix-go/services/blockreferencetime"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"math"
	"time"
)

const CallCommitteeContractInterval = 200 * time.Millisecond

type LeanHelixTerm struct {
	*ConsensusMessagesFilter
	termInCommittee *termincommittee.TermInCommittee
}

func NewLeanHelixTerm(ctx context.Context, logger logger.LHLogger, config *interfaces.Config, state *state.State, electionTrigger interfaces.ElectionScheduler, onCommit interfaces.OnCommitCallback, prevBlock interfaces.Block, prevBlockProofBytes []byte, canBeFirstLeader bool) *LeanHelixTerm {
	prevBlockProof := protocol.BlockProofReader(prevBlockProofBytes)
	randomSeed := randomseed.CalculateRandomSeed(prevBlockProof.RandomSeedSignature())
	blockHeight := blockheight.GetBlockHeight(prevBlock) + 1
	prevBlockRefTime := blockreferencetime.GetBlockReferenceTime(prevBlock)
	myMemberId := config.Membership.MyMemberId()
	messageFactory := messagesfactory.NewMessageFactory(config.InstanceId, config.KeyManager, myMemberId, randomSeed)

	committeeMembers, err := requestOrderedCommitteePersist(state, blockHeight, randomSeed, prevBlockRefTime, config, logger)
	if err != nil {
		logger.Info("ERROR RECEIVING COMMITTEE: H=%d, error=%s", blockHeight, err)
	}
	// on ctx terminated requestOrderedCommitteePersist returns nil committee
	isParticipating := isParticipatingInTerm(myMemberId, committeeMembers)

	if !isParticipating {
		logger.Debug("OUT OF COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, members=%s, isParticipating=%t", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, termincommittee.ToCommitteeMembersStr(committeeMembers), isParticipating)
		return termNotInCommittee(randomSeed, config)
	}

	logger.Debug("RECEIVED COMMITTEE: H=%d, prevBlockProof=%s, randomSeed=%d, refTime=%d, members=%s, isParticipating=%t", blockHeight, printShortBlockProofBytes(prevBlockProofBytes), randomSeed, prevBlockRefTime, termincommittee.ToCommitteeMembersStr(committeeMembers), isParticipating)
	logger.ConsensusTrace("got committee for the current consensus round", nil, log.StringableSlice("committee", termincommittee.GetMemberIds(committeeMembers)))

	termInCommittee := termincommittee.NewTermInCommittee(logger, config, state, messageFactory, electionTrigger, committeeMembers, prevBlock, canBeFirstLeader, CommitsToProof(logger, config.KeyManager, onCommit))
	return &LeanHelixTerm{
		ConsensusMessagesFilter: NewConsensusMessagesFilter(termInCommittee, config.KeyManager, randomSeed),
		termInCommittee:         termInCommittee,
	}
}

func requestOrderedCommitteePersist(s *state.State, blockHeight primitives.BlockHeight, randomSeed uint64, prevBlockReferenceTime primitives.TimestampSeconds, config *interfaces.Config, logger logger.LHLogger) ([]interfaces.CommitteeMember, error) {
	const maxView = primitives.View(math.MaxUint64)
	ctx, err := s.Contexts.For(state.NewHeightView(blockHeight, maxView)) // term-level context
	if err != nil {
		return nil, err
	}
	logger.Debug("Polling RequestOrderedCommittee: H=%d, interval-between-attempts=%d", blockHeight, CallCommitteeContractInterval)

	attempts := 1
	for {

		// exit on term update (node sync) or system shutdown
		if ctx.Err() != nil {
			return nil, errors.Wrap(ctx.Err(), "requestOrderedCommitteePersist: context terminated")
		}

		committeeMembers, err := config.Membership.RequestOrderedCommittee(ctx, blockHeight, randomSeed, prevBlockReferenceTime)
		if err == nil {
			return committeeMembers, nil
		}

		// log every 500 failures
		if attempts%500 == 1 {
			if ctx.Err() == nil { // this may fail rightfully on graceful shutdown (ctx.Done), we don't want to report an error in this case
				logger.Info("requestOrderedCommitteePersist: cannot get ordered committee #attempts=%d, error=%s", attempts, err)
			}
		}

		// sleep or wait for ctx done, whichever comes first
		sleepOrShutdown, cancel := context.WithTimeout(ctx, CallCommitteeContractInterval)
		<-sleepOrShutdown.Done()
		cancel()

		attempts++
	}
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

func isParticipatingInTerm(myMemberId primitives.MemberId, committeeMembers []interfaces.CommitteeMember) bool {
	for _, committeeMember := range committeeMembers {
		if myMemberId.Equal(committeeMember.Id) {
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
