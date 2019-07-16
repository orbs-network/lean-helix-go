// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/leanhelixterm"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/services/rawmessagesfilter"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
	"math"
)

type blockWithProof struct {
	block               interfaces.Block
	prevBlockProofBytes []byte
}

type MessageWithContext struct {
	msg *interfaces.ConsensusRawMessage
	ctx context.Context
}

type WorkerLoop struct {
	MessagesChannel    chan *MessageWithContext
	UpdateStateChannel chan *blockWithProof
	ElectionChannel    chan func(ctx context.Context)
	electionTrigger    interfaces.ElectionTrigger

	currentHeight               primitives.BlockHeight
	config                      *interfaces.Config
	logger                      L.LHLogger
	filter                      *rawmessagesfilter.RawMessageFilter
	leanHelixTerm               *leanhelixterm.LeanHelixTerm
	onCommitCallback            interfaces.OnCommitCallback
	onNewConsensusRoundCallback interfaces.OnNewConsensusRoundCallback
	onUpdateStateCallback       interfaces.OnUpdateStateCallback
}

func NewWorkerLoop(config *interfaces.Config, logger L.LHLogger, electionTrigger interfaces.ElectionTrigger, onCommitCallback interfaces.OnCommitCallback) *WorkerLoop {

	logger.Debug(L.LC(math.MaxUint64, math.MaxUint64, config.Membership.MyMemberId()), "LHFLOW NewWorkerLoop()")
	filter := rawmessagesfilter.NewConsensusMessageFilter(config.InstanceId, config.Membership.MyMemberId(), logger)
	return &WorkerLoop{
		MessagesChannel:    make(chan *MessageWithContext, 10),
		UpdateStateChannel: make(chan *blockWithProof),
		ElectionChannel:    make(chan func(ctx context.Context)),
		electionTrigger:    electionTrigger,
		currentHeight:      0,
		config:             config,
		logger:             logger,
		filter:             filter,
		onCommitCallback:   onCommitCallback,
	}
}

func (lh *WorkerLoop) Run(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()
	lh.logger.Debug(L.LC(math.MaxUint64, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW WORKERLOOP START")
	lh.logger.Debug(L.LC(math.MaxUint64, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHMSG WORKERLOOP START LISTENING NOW")
	for {
		lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW WORKERLOOP LISTENING")

		select {
		case <-ctx.Done(): // system shutdown
			lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW WORKERLOOP DONE, Terminating Run().")
			lh.logger.Debug(L.LC(math.MaxUint64, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHMSG WORKERLOOP STOPPED LISTENING")
			return

		case res := <-lh.MessagesChannel:
			lh.filter.HandleConsensusRawMessage(res.ctx, res.msg)

		case trigger := <-lh.ElectionChannel:
			if trigger == nil {
				// this cannot happen, ignore
				lh.logger.Info(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "XXXXXX LHFLOW WORKERLOOP ELECTION, OMG trigger is nil, not triggering election!")
			}
			lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW WORKERLOOP ELECTION")
			trigger(ctx)

		case receivedBlockWithProof := <-lh.UpdateStateChannel: // NodeSync
			lh.HandleUpdateState(ctx, receivedBlockWithProof)
		}
	}
}

func (lh *WorkerLoop) HandleUpdateState(ctx context.Context, receivedBlockWithProof *blockWithProof) {
	receivedBlockHeight := blockheight.GetBlockHeight(receivedBlockWithProof.block)
	if receivedBlockHeight >= lh.currentHeight {
		lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW WORKERLOOP UPDATESTATE ACCEPTED block with height=%d, calling onNewConsensusRound()", receivedBlockHeight)
		// This block is received from external source
		// Refuse to be leader on V=0 for a block received from block sync, because this block will usually be not be the latest block.
		lh.onNewConsensusRound(ctx, receivedBlockWithProof.block, receivedBlockWithProof.prevBlockProofBytes, false)
	} else {
		lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW WORKERLOOP UPDATESTATE IGNORE - Received block ignored because its height=%d is less than current height=%d", receivedBlockHeight, lh.currentHeight)
	}
}

func GetMemberIdsFromBlockProof(blockProofBytes []byte) ([]primitives.MemberId, error) {
	if blockProofBytes == nil || len(blockProofBytes) == 0 {
		return nil, errors.Errorf("GetMemberIdsFromBlockProof: nil blockProof - cannot deduce members locally")
	}
	blockProof := protocol.BlockProofReader(blockProofBytes)
	sendersIterator := blockProof.NodesIterator()
	committeeMembers := make([]primitives.MemberId, 0)
	for sendersIterator.HasNext() {
		committeeMembers = append(committeeMembers, sendersIterator.NextNodes().MemberId())
	}
	return committeeMembers, nil
}

func (lh *WorkerLoop) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, maybePrevBlockProofBytes []byte) error {

	if block == nil {
		return errors.Errorf("ValidateBlockConsensus: nil block")
	}
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "ValidateBlockConsensus START for blockHeight=%d", block.Height())
	if blockProofBytes == nil || len(blockProofBytes) == 0 {
		return errors.Errorf("ValidateBlockConsensus: nil blockProof")
	}

	blockProof := protocol.BlockProofReader(blockProofBytes)
	blockRefFromProof := blockProof.BlockRef()
	if blockRefFromProof.MessageType() != protocol.LEAN_HELIX_COMMIT {
		return errors.Errorf("ValidateBlockConsensus: Message is not COMMIT, it is %v", blockRefFromProof.MessageType())
	}

	if lh.config.InstanceId != blockRefFromProof.InstanceId() {
		return errors.Errorf("ValidateBlockConsensus: Mismatched InstanceID: config=%v blockProof=%v", lh.config.InstanceId, blockRefFromProof.InstanceId())
	}

	blockHeight := block.Height()
	if blockHeight != blockRefFromProof.BlockHeight() {
		return errors.Errorf("ValidateBlockConsensus: Mismatched block height: block=%v blockProof=%v", blockHeight, block.Height())
	}

	if !lh.config.BlockUtils.ValidateBlockCommitment(blockHeight, block, blockRefFromProof.BlockHash()) {
		return errors.Errorf("ValidateBlockConsensus: ValidateBlockCommitment() failed")
	}

	// note: it is ok to disregard the order of committee here (hence randomSeed is not calculated) - the blockProof only checks for set of quorum COMMITS
	committeeMembers, err := lh.config.Membership.RequestOrderedCommittee(ctx, blockHeight, 0)
	if err != nil { // support for failure in committee calculation
		return err
	}
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "ValidateBlockConsensus: RECEIVED COMMITTEE for H=%d, members=%s", blockHeight, termincommittee.ToCommitteeMembersStr(committeeMembers))

	sendersIterator := blockProof.NodesIterator()
	set := make(map[storage.MemberIdStr]bool)
	var sendersCounter = 0
	for {
		if !sendersIterator.HasNext() {
			break
		}

		sender := sendersIterator.NextNodes()
		if err := proofsvalidator.VerifyBlockRefMessage(blockRefFromProof, sender, lh.config.KeyManager); err != nil {
			return errors.Wrapf(err, "ValidateBlockConsensus: VerifyBlockRefMessage() failed")
		}

		memberId := sender.MemberId()
		if _, ok := set[storage.MemberIdStr(memberId)]; ok {
			return errors.Errorf("ValidateBlockConsensus: Could not read memberId=%s from set", termincommittee.Str(memberId))
		}

		if !proofsvalidator.IsInMembers(committeeMembers, memberId) {
			return errors.Errorf("ValidateBlockConsensus: Id=%s which signed block with H=%d is not part of committee of that block height. Committee=%s", termincommittee.Str(memberId), blockHeight, termincommittee.ToCommitteeMembersStr(committeeMembers))
		}

		set[storage.MemberIdStr(memberId)] = true
		sendersCounter++
	}

	q := quorum.CalcQuorumSize(len(committeeMembers))
	if sendersCounter < q {
		return errors.Errorf("ValidateBlockConsensus: sendersCounter=%d is less than quorum=%d (committeeMembersCount=%d)", sendersCounter, q, len(committeeMembers))
	}

	if len(blockProof.RandomSeedSignature()) == 0 || blockProof.RandomSeedSignature() == nil {
		return errors.Errorf("ValidateBlockConsensus: blockProof does not contain randomSeed")
	}

	prevBlockProof := protocol.BlockProofReader(maybePrevBlockProofBytes)
	if err := randomseed.ValidateRandomSeed(lh.config.KeyManager, blockHeight, blockProof, prevBlockProof); err != nil {
		return errors.Wrapf(err, "ValidateBlockConsensus: ValidateRandomSeed() failed")
	}
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "ValidateBlockConsensus PASSED for blockHeight=%s", block.Height())

	return nil
}

// TODO Is this for testing only? maybe it shouldn't be here
func (lh *WorkerLoop) HandleConsensusMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {
	select {
	case <-ctx.Done():
		lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "HandleConsensusRawMessage() ID=%s CONTEXT TERMINATED", termincommittee.Str(lh.config.Membership.MyMemberId()))
		return

	case lh.MessagesChannel <- &MessageWithContext{ctx: ctx, msg: message}:
	}
}

func (lh *WorkerLoop) onCommit(ctx context.Context, block interfaces.Block, blockProofBytes []byte) {
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW onCommitCallback START from leanhelix.onCommit()")
	lh.onCommitCallback(ctx, block, blockProofBytes)
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW onCommitCallback RETURNED from leanhelix.onCommit()")
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "Calling onNewConsensusRound() from leanhelix.onCommit()")
	lh.onNewConsensusRound(ctx, block, blockProofBytes, true)
}

func (lh *WorkerLoop) onNewConsensusRound(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte, canBeFirstLeader bool) {
	lh.currentHeight = blockheight.GetBlockHeight(prevBlock) + 1
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "onNewConsensusRound() INCREMENTED HEIGHT TO %d", lh.currentHeight)
	if lh.leanHelixTerm != nil {
		lh.leanHelixTerm.Dispose()
		lh.leanHelixTerm = nil
	}
	lh.leanHelixTerm = leanhelixterm.NewLeanHelixTerm(ctx, lh.logger, lh.config, lh.electionTrigger, lh.onCommit, prevBlock, prevBlockProofBytes, canBeFirstLeader)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
	if lh.onNewConsensusRoundCallback != nil {
		lh.onNewConsensusRoundCallback(ctx, prevBlock, canBeFirstLeader)
	}
}

func (lh *WorkerLoop) GetCurrentHeight() primitives.BlockHeight {
	return lh.currentHeight
}
