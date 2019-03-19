package leanhelix

import (
	"context"
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

type LeanHelix struct {
	messagesChannel    chan *interfaces.ConsensusRawMessage
	updateStateChannel chan *blockWithProof
	currentHeight      primitives.BlockHeight
	config             *interfaces.Config
	logger             L.LHLogger
	filter             *rawmessagesfilter.RawMessageFilter
	leanHelixTerm      *leanhelixterm.LeanHelixTerm
	onCommitCallback   interfaces.OnCommitCallback
}

// ***********************************
// LeanHelix Constructor
// ***********************************
func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback) *LeanHelix {
	var lhLog L.LHLogger
	if config.Logger == nil {
		lhLog = L.NewLhLogger(L.NewSilentLogger())
	} else {
		lhLog = L.NewLhLogger(config.Logger)
	}

	lhLog.Debug(L.LC(math.MaxUint64, math.MaxUint64, config.Membership.MyMemberId()), "LHFLOW NewLeanHelix()")
	filter := rawmessagesfilter.NewConsensusMessageFilter(config.InstanceId, config.Membership.MyMemberId(), lhLog)
	return &LeanHelix{
		messagesChannel:    make(chan *interfaces.ConsensusRawMessage),
		updateStateChannel: make(chan *blockWithProof),
		currentHeight:      0,
		config:             config,
		logger:             lhLog,
		filter:             filter,
		onCommitCallback:   onCommitCallback,
	}
}

func (lh *LeanHelix) Run(ctx context.Context) {
	lh.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW MAINLOOP START")
	lh.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHMSG START LISTENING NOW")
	for {
		select {
		case <-ctx.Done():
			lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW MAINLOOP DONE, Terminating Run().")
			lh.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHMSG STOPPED LISTENING")
			return

		case message := <-lh.messagesChannel:
			lh.filter.HandleConsensusRawMessage(ctx, message)

		case trigger := <-lh.config.ElectionTrigger.ElectionChannel():
			if trigger == nil {
				// this cannot happen, ignore
				lh.logger.Info(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "XXXXXX LHFLOW MAINLOOP ELECTION, OMG trigger is nil, not triggering election!")
			}
			lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW MAINLOOP ELECTION")
			trigger(ctx)

		case receivedBlockWithProof := <-lh.updateStateChannel:
			receivedBlockHeight := blockheight.GetBlockHeight(receivedBlockWithProof.block)
			if receivedBlockHeight >= lh.currentHeight {
				lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW MAINLOOP UPDATESTATE ACCEPTED block with height=%d, calling onNewConsensusRound()", receivedBlockHeight)
				lh.onNewConsensusRound(ctx, receivedBlockWithProof.block, receivedBlockWithProof.prevBlockProofBytes, false)
			} else {
				lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW MAINLOOP UPDATESTATE IGNORE - Received block ignored because its height=%d is less than current height=%d", receivedBlockHeight, lh.currentHeight)
			}
		}
	}
}

func (lh *LeanHelix) UpdateState(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {
	var prevBlockHeight primitives.BlockHeight
	if prevBlock != nil {
		prevBlockHeight = prevBlock.Height()
	}
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW MAINLOOP UPDATESTATE Writing to UPDATESTATE channel, prevBlockHeight=%d", prevBlockHeight)
	select {
	case <-ctx.Done():
		lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW MAINLOOP UPDATESTATE DONE")
		return

	case lh.updateStateChannel <- &blockWithProof{prevBlock, prevBlockProofBytes}:
		return
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

func (lh *LeanHelix) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, maybePrevBlockProofBytes []byte) error {

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
	lh.logger.Info(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "ValidateBlockConsensus: RECEIVED COMMITTEE for H=%d, members=%s", blockHeight, termincommittee.ToCommitteeMembersStr(committeeMembers))

	sendersIterator := blockProof.NodesIterator()
	set := make(map[storage.MemberIdStr]bool)
	var sendersCounter = 0
	for {
		if !sendersIterator.HasNext() {
			break
		}

		sender := sendersIterator.NextNodes()
		if !proofsvalidator.VerifyBlockRefMessage(blockRefFromProof, sender, lh.config.KeyManager) {
			return errors.Errorf("ValidateBlockConsensus: VerifyBlockRefMessage() failed")
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

func (lh *LeanHelix) HandleConsensusMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {
	select {
	case <-ctx.Done():
		lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "HandleConsensusRawMessage() ID=%s CONTEXT TERMINATED", termincommittee.Str(lh.config.Membership.MyMemberId()))
		return

	case lh.messagesChannel <- message:
	}
}

func (lh *LeanHelix) onCommit(ctx context.Context, block interfaces.Block, blockProofBytes []byte) {
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW onCommitCallback START from leanhelix.onCommit()")
	lh.onCommitCallback(ctx, block, blockProofBytes)
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "LHFLOW onCommitCallback RETURNED from leanhelix.onCommit()")
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "Calling onNewConsensusRound from leanhelix.onCommit()")
	lh.onNewConsensusRound(ctx, block, blockProofBytes, true)
}

func (lh *LeanHelix) onNewConsensusRound(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte, canBeFirstLeader bool) {
	lh.currentHeight = blockheight.GetBlockHeight(prevBlock) + 1
	lh.logger.Debug(L.LC(lh.currentHeight, math.MaxUint64, lh.config.Membership.MyMemberId()), "onNewConsensusRound() INCREMENTED HEIGHT TO %d", lh.currentHeight)
	if lh.leanHelixTerm != nil {
		lh.leanHelixTerm.Dispose()
		lh.leanHelixTerm = nil
	}
	lh.leanHelixTerm = leanhelixterm.NewLeanHelixTerm(ctx, lh.logger, lh.config, lh.onCommit, prevBlock, prevBlockProofBytes, canBeFirstLeader)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
}
