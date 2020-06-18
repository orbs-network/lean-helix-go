// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/blockreferencetime"
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
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/pkg/errors"
)

type blockWithProof struct {
	block               interfaces.Block
	prevBlockProofBytes []byte
}

type WorkerLoop struct {
	MessagesChannel             chan *interfaces.ConsensusRawMessage
	workerUpdateStateChannel    chan *blockWithProof
	electionChannel             chan *interfaces.ElectionTrigger
	electionTrigger             interfaces.ElectionScheduler
	state                       *state.State
	config                      *interfaces.Config
	logger                      L.LHLogger
	filter                      *rawmessagesfilter.RawMessageFilter
	leanHelixTerm               *leanhelixterm.LeanHelixTerm
	onCommitCallback            interfaces.OnCommitCallback
	onNewConsensusRoundCallback interfaces.OnNewConsensusRoundCallback
}

func NewWorkerLoop(
	state *state.State,
	config *interfaces.Config,
	logger L.LHLogger,
	electionTrigger interfaces.ElectionScheduler,
	onCommitCallback interfaces.OnCommitCallback,
	onNewConsensusRoundCallback interfaces.OnNewConsensusRoundCallback) *WorkerLoop {

	logger.Debug("LHFLOW NewWorkerLoop()")
	filter := rawmessagesfilter.NewConsensusMessageFilter(config.InstanceId, config.Membership.MyMemberId(), logger, state)
	return &WorkerLoop{
		MessagesChannel:             make(chan *interfaces.ConsensusRawMessage, 1000), // TODO config.MsgChanBufLen
		workerUpdateStateChannel:    make(chan *blockWithProof, 1),                    // must be at least 1 // TODO config.UpdateStateChanBufLen
		electionChannel:             make(chan *interfaces.ElectionTrigger, 1),        // must be at least 1 // TODO config.ElectionChanBufLen
		electionTrigger:             electionTrigger,
		state:                       state,
		config:                      config,
		logger:                      logger,
		filter:                      filter,
		onCommitCallback:            onCommitCallback,
		onNewConsensusRoundCallback: onNewConsensusRoundCallback,
	}
}

func (lh *WorkerLoop) Run(ctx context.Context) {
	lh.logger.Debug("LHFLOW LHMSG WORKERLOOP START LISTENING NOW")
	for {
		select {
		case <-ctx.Done(): // system shutdown
			lh.logger.Info("LHFLOW WORKERLOOP DONE STOPPED LISTENING, SHUTDOWN START")
			lh.cleanupCurrentTerm()
			lh.logger.Info("LHFLOW WORKERLOOP DONE STOPPED LISTENING, SHUTDOWN END")
			return

		case msg := <-lh.MessagesChannel:
			parsedMessage := interfaces.ToConsensusMessage(msg)
			lh.logger.Debug("LHFLOW LHMSG WORKERLOOP RECEIVED %v from %v for H=%d V=%d", parsedMessage.MessageType(), parsedMessage.SenderMemberId(), parsedMessage.BlockHeight(), parsedMessage.View())
			lh.filter.HandleConsensusRawMessage(msg)

		case trigger := <-lh.electionChannel:
			if trigger == nil {
				// this cannot happen, ignore
				lh.logger.Info("XXXXXX LHFLOW WORKERLOOP ELECTION, OMG trigger is nil, not triggering election!")
				continue
			}
			current := lh.state.HeightView()
			if current.Height() != trigger.Hv.Height() || current.View() != trigger.Hv.View() { // stale election message
				lh.logger.Info("LHFLOW WORKERLOOP ELECTION - INVALID HEIGHT/VIEW IGNORED - Current: %s, ElectionTrigger: %s",
					current, trigger.Hv)
				continue
			}

			lh.logger.Debug("LHFLOW WORKERLOOP ELECTION")
			trigger.MoveToNextLeader()

		case receivedBlockWithProof := <-lh.workerUpdateStateChannel: // NodeSync
			var height primitives.BlockHeight

			if receivedBlockWithProof.block != nil {
				height = receivedBlockWithProof.block.Height()
			}
			lh.logger.Debug("LHFLOW UPDATESTATE WORKERLOOP - Received block with H=%d", height)
			lh.handleUpdateState(receivedBlockWithProof)
			lh.logger.Debug("LHFLOW UPDATESTATE WORKERLOOP - Handled block with H=%d", height)
		}
	}
}

func (lh *WorkerLoop) handleUpdateState(receivedBlockWithProof *blockWithProof) {
	receivedBlockHeight := blockheight.GetBlockHeight(receivedBlockWithProof.block)

	if receivedBlockHeight >= lh.state.Height() {
		lh.logger.Debug("LHFLOW UPDATESTATE WORKERLOOP ACCEPTED block with height=%d, calling onNewConsensusRound() from handleUpdateState", receivedBlockHeight)
		// This block is received from external source
		// Refuse to be leader on V=0 for a block received from block sync, because this block will usually be not be the latest block.
		lh.onNewConsensusRound(receivedBlockWithProof.block, receivedBlockWithProof.prevBlockProofBytes, false)
	} else {
		lh.logger.Debug("LHFLOW UPDATESTATE WORKERLOOP IGNORE - Received block ignored because its height=%d is less than current height=%d", receivedBlockHeight, lh.state.Height())
	}
}

func (lh *WorkerLoop) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, prevBlock interfaces.Block, maybePrevBlockProofBytes []byte) error {
	if ctx.Err() != nil {
		return errors.New("context canceled")
	}
	if block == nil {
		return errors.Errorf("ValidateBlockConsensus: nil block")
	}
	lh.logger.Debug("ValidateBlockConsensus START for blockHeight=%d", block.Height())
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
		return errors.Errorf("ValidateBlockConsensus: Mismatched height: blockHeight=%v but blockProof.height=%v", blockHeight, blockRefFromProof.BlockHeight())
	}

	if !lh.config.BlockUtils.ValidateBlockCommitment(blockHeight, block, blockRefFromProof.BlockHash()) {
		return errors.Errorf("ValidateBlockConsensus: ValidateBlockCommitment() failed")
	}

	// note: it is ok to disregard the order of committee here (hence randomSeed is not calculated) - the blockProof only checks for set of quorum COMMITS
	committeeMembers, committeeWeights, err := lh.config.Membership.RequestCommitteeForBlockProof(ctx, blockreferencetime.GetBlockReferenceTime(prevBlock))
	if err != nil { // support for failure in committee calculation
		return err
	}
	lh.logger.Info("ValidateBlockConsensus: RECEIVED COMMITTEE for H=%d, RefTime=%d, members=%s", blockHeight, blockreferencetime.GetBlockReferenceTime(prevBlock), termincommittee.ToCommitteeMembersStr(committeeMembers))

	sendersIterator := blockProof.NodesIterator()
	set := make(map[storage.MemberIdStr]bool)
	sendersTotalWeight := uint(0)
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

		// todo this logic is repeated
		for i := 0; i < len(committeeMembers); i++ {
			if committeeMembers[i].Equal(memberId) {
				sendersTotalWeight += committeeWeights[i]
				break
			}
		} // todo what if not found?
	}

	// todo this logic is repeated
	q := quorum.CalcQuorumWeight(committeeWeights)
	if sendersTotalWeight < q {
		return errors.Errorf("ValidateBlockConsensus: sendersTotalWeight=%d is less than quorum=%d (committeeMembersCount=%d)", sendersTotalWeight, q, len(committeeMembers))
	}

	if len(blockProof.RandomSeedSignature()) == 0 || blockProof.RandomSeedSignature() == nil {
		return errors.Errorf("ValidateBlockConsensus: blockProof does not contain randomSeed")
	}

	prevBlockProof := protocol.BlockProofReader(maybePrevBlockProofBytes)
	if err := randomseed.ValidateRandomSeed(lh.config.KeyManager, blockHeight, blockProof, prevBlockProof); err != nil {
		return errors.Wrapf(err, "ValidateBlockConsensus: ValidateRandomSeed() failed")
	}
	lh.logger.Debug("ValidateBlockConsensus PASSED for blockHeight=%s", block.Height())

	return nil
}

func (lh *WorkerLoop) onCommit(ctx context.Context, block interfaces.Block, blockProofBytes []byte) error {
	height := block.Height()
	lh.logger.Debug("LHFLOW onCommitCallback START from leanhelix.onCommit() ID=%s H=%d", lh.config.Membership.MyMemberId(), height)

	err := lh.onCommitCallback(ctx, block, blockProofBytes)
	lh.logger.ConsensusTrace("sent block to commit callback", err)
	if err != nil {
		lh.logger.Debug("LHFLOW onCommitCallback FAILED - %s", err.Error())
		return err
	}
	lh.logger.Debug("LHFLOW onCommitCallback RETURNED from leanhelix.onCommit()")
	lh.logger.Debug("Calling onNewConsensusRound() from leanhelix.onCommit()")
	lh.onNewConsensusRound(block, blockProofBytes, true)

	return nil
}

func (lh *WorkerLoop) onNewConsensusRound(prevBlock interfaces.Block, prevBlockProofBytes []byte, canBeFirstLeader bool) {
	hv := state.NewHeightView(blockheight.GetBlockHeight(prevBlock)+1, 0)
	ctx, err := lh.state.Contexts.For(hv)
	if err != nil {
		lh.logger.Info("onNewConsensusRound() error: %e", err)
		return
	}

	current, err := lh.state.SetHeightAndResetView(hv.Height())
	if err != nil {
		lh.logger.Info("onNewConsensusRound() failed height increment %d: %s", current.Height(), err)
		return
	}

	lh.logger.Debug("onNewConsensusRound() INCREMENTED HEIGHT TO %d", current.Height())
	if lh.leanHelixTerm != nil {
		lh.leanHelixTerm.Dispose()
		lh.leanHelixTerm = nil
	}

	lh.logger.ConsensusTrace("starting a new consensus round", nil)

	lh.leanHelixTerm = leanhelixterm.NewLeanHelixTerm(ctx, lh.logger, lh.config, lh.state, lh.electionTrigger, lh.onCommit, prevBlock, prevBlockProofBytes, canBeFirstLeader)
	lh.logger.Debug("onNewConsensusRound() Calling ConsumeCacheMessages for H=%d", lh.state.Height())
	lh.filter.ConsumeCacheMessages(lh.leanHelixTerm)
	if lh.onNewConsensusRoundCallback != nil {
		lh.onNewConsensusRoundCallback(ctx, lh.state.Height(), prevBlock, canBeFirstLeader)
	}
}

func (lh *WorkerLoop) cleanupCurrentTerm() {
	if lh.leanHelixTerm != nil {
		lh.leanHelixTerm.Dispose()
	}
}

func (lh *WorkerLoop) interrupt() {
	lh.state.Contexts.Shutdown()
}
