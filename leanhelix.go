package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/leanhelixterm"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/services/rawmessagesfilter"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
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
	logger             interfaces.Logger
	filter             *rawmessagesfilter.RawMessageFilter
	leanHelixTerm      *leanhelixterm.LeanHelixTerm
	onCommitCallback   interfaces.OnCommitCallback
}

// ***********************************
// LeanHelix Constructor
// ***********************************
func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback) *LeanHelix {
	if config.Logger == nil {
		config.Logger = logger.NewSilentLogger()
	}

	config.Logger.Debug("NewLeanHelix() ID=%s", termincommittee.Str(config.Membership.MyMemberId()))
	filter := rawmessagesfilter.NewConsensusMessageFilter(config.InstanceId, config.Membership.MyMemberId(), config.Logger)
	return &LeanHelix{
		messagesChannel:    make(chan *interfaces.ConsensusRawMessage),
		updateStateChannel: make(chan *blockWithProof),
		currentHeight:      0,
		config:             config,
		logger:             config.Logger,
		filter:             filter,
		onCommitCallback:   onCommitCallback,
	}
}

func (lh *LeanHelix) Run(ctx context.Context) {
	lh.logger.Info("H=X V=X ID=%s LHFLOW Run() Starting infinite loop", termincommittee.Str(lh.config.Membership.MyMemberId()))
	for {
		select {
		case <-ctx.Done():
			lh.logger.Debug("H=%d V=X ID=%s LHFLOW Run() Received <Done>. Terminating Run().", lh.currentHeight, termincommittee.Str(lh.config.Membership.MyMemberId()))
			return

		case message := <-lh.messagesChannel:
			lh.filter.HandleConsensusRawMessage(ctx, message)

		case trigger := <-lh.config.ElectionTrigger.ElectionChannel():
			if trigger == nil {
				// this cannot happen, ignore
				lh.logger.Info("H=%d V=X ID=%s XXXXXX LHFLOW Run() Election, OMG trigger is nil!", lh.currentHeight, termincommittee.Str(lh.config.Membership.MyMemberId()))
			}
			lh.logger.Debug("H=%d V=X ID=%s LHFLOW Run() Received <Election>", lh.currentHeight, termincommittee.Str(lh.config.Membership.MyMemberId()))
			trigger(ctx)

		case receivedBlockWithProof := <-lh.updateStateChannel:
			receivedBlockHeight := blockheight.GetBlockHeight(receivedBlockWithProof.block)
			if receivedBlockHeight >= lh.currentHeight {
				lh.logger.Debug("H=%d V=X ID=%s LHFLOW Run() Received <UpdateState> Calling onNewConsensusRound() receivedBlockHeight=%d", lh.currentHeight, termincommittee.Str(lh.config.Membership.MyMemberId()), receivedBlockHeight)
				lh.onNewConsensusRound(ctx, receivedBlockWithProof.block, receivedBlockWithProof.prevBlockProofBytes)
			} else {
				lh.logger.Debug("H=%d V=X ID=%s LHFLOW Run() Ignoring received block because its height=%d is less than current height=%d", lh.currentHeight, termincommittee.Str(lh.config.Membership.MyMemberId()), receivedBlockHeight, lh.currentHeight)
			}
		}
	}
}

func (lh *LeanHelix) UpdateState(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {
	select {
	case <-ctx.Done():
		return

	case lh.updateStateChannel <- &blockWithProof{prevBlock, prevBlockProofBytes}:
		return
	}

}

func (lh *LeanHelix) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, prevBlockProofBytes []byte) error {

	if block == nil {
		return errors.Errorf("ValidateBlockConsensus(): nil block")
	}
	lh.logger.Debug("ValidateBlockConsensus() ID=%s HEIGHT=%s", termincommittee.Str(lh.config.Membership.MyMemberId()), block.Height())
	if blockProofBytes == nil || len(blockProofBytes) == 0 {
		return errors.Errorf("ValidateBlockConsensus(): nil blockProof")
	}

	blockProof := protocol.BlockProofReader(blockProofBytes)
	blockRefFromProof := blockProof.BlockRef()
	if blockRefFromProof.MessageType() != protocol.LEAN_HELIX_COMMIT {
		return errors.Errorf("ValidateBlockConsensus(): Message is not COMMIT, it is %v", blockRefFromProof.MessageType())
	}

	if lh.config.InstanceId != blockRefFromProof.InstanceId() {
		return errors.Errorf("ValidateBlockConsensus(): Mismatched InstanceID: config=%v blockProof=%v", lh.config.InstanceId, blockRefFromProof.InstanceId())
	}

	blockHeight := block.Height()
	if blockHeight != blockRefFromProof.BlockHeight() {
		return errors.Errorf("ValidateBlockConsensus(): Mismatched block height: block=%v blockProof=%v", blockHeight, block.Height())
	}

	if !lh.config.BlockUtils.ValidateBlockCommitment(blockHeight, block, blockRefFromProof.BlockHash()) {
		return errors.Errorf("ValidateBlockConsensus(): ValidateBlockCommitment() failed")
	}

	committeeMembers := lh.config.Membership.RequestOrderedCommittee(ctx, blockHeight, 0)

	sendersIterator := blockProof.NodesIterator()
	set := make(map[storage.MemberIdStr]bool)
	var sendersCounter = 0
	for {
		if !sendersIterator.HasNext() {
			break
		}

		sender := sendersIterator.NextNodes()
		if !proofsvalidator.VerifyBlockRefMessage(blockRefFromProof, sender, lh.config.KeyManager) {
			return errors.Errorf("ValidateBlockConsensus(): VerifyBlockRefMessage() failed")
		}

		memberId := sender.MemberId()
		if _, ok := set[storage.MemberIdStr(memberId)]; ok {
			return errors.Errorf("ValidateBlockConsensus(): Could not read memberId=%s from set", storage.MemberIdStr(memberId))
		}

		if !proofsvalidator.IsInMembers(committeeMembers, memberId) {
			return errors.Errorf("ValidateBlockConsensus(): memberId=%s is not part of committee", storage.MemberIdStr(memberId))
		}

		set[storage.MemberIdStr(memberId)] = true
		sendersCounter++
	}

	q := quorum.CalcQuorumSize(len(committeeMembers))
	if sendersCounter < q {
		return errors.Errorf("ValidateBlockConsensus(): sendersCounter=%d is less that quorum=%d", sendersCounter, q)
	}

	if len(blockProof.RandomSeedSignature()) == 0 || blockProof.RandomSeedSignature() == nil {
		return errors.Errorf("ValidateBlockConsensus(): blockProof does not contain randomSeed")
	}

	prevBlockProof := protocol.BlockProofReader(prevBlockProofBytes)
	if err := randomseed.ValidateRandomSeed(lh.config.KeyManager, blockHeight, blockProof, prevBlockProof); err != nil {
		return errors.Wrapf(err, "ValidateBlockConsensus(): ValidateRandomSeed() failed")
	}

	return nil
}

func (lh *LeanHelix) HandleConsensusMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {
	lh.logger.Debug("HandleConsensusRawMessage() ID=%s", termincommittee.Str(lh.config.Membership.MyMemberId()))
	select {
	case <-ctx.Done():
		return

	case lh.messagesChannel <- message:
	}
}

// ************************ Internal ***************************************

func (lh *LeanHelix) onCommit(ctx context.Context, block interfaces.Block, blockProofBytes []byte) {
	lh.onCommitCallback(ctx, block, blockProofBytes)
	lh.logger.Debug("Calling onNewConsensusRound() from onCommit() lh.currentHeight=%d", lh.currentHeight)
	lh.onNewConsensusRound(ctx, block, blockProofBytes)
}

func (lh *LeanHelix) onNewConsensusRound(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {
	lh.currentHeight = blockheight.GetBlockHeight(prevBlock) + 1
	lh.leanHelixTerm = leanhelixterm.NewLeanHelixTerm(ctx, lh.config, lh.onCommit, prevBlock, prevBlockProofBytes)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
}
