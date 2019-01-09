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
)

type blockWithProof struct {
	block               interfaces.Block
	prevBlockProofBytes []byte
}

type LeanHelix struct {
	messagesChannel         chan *interfaces.ConsensusRawMessage
	acknowledgeBlockChannel chan *blockWithProof
	currentHeight           primitives.BlockHeight
	config                  *interfaces.Config
	logger                  interfaces.Logger
	filter                  *rawmessagesfilter.RawMessageFilter
	leanHelixTerm           *leanhelixterm.LeanHelixTerm
	onCommitCallback        interfaces.OnCommitCallback
}

// ***********************************
// LeanHelix Constructor
// ***********************************
func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback) *LeanHelix {
	if config.Logger == nil {
		config.Logger = logger.NewSilentLogger()
	}

	config.Logger.Debug("NewLeanHelix() ID=%s", termincommittee.Str(config.Membership.MyMemberId()))
	filter := rawmessagesfilter.NewRawMessageFilter(config.InstanceId, config.Membership.MyMemberId(), config.Logger)
	return &LeanHelix{
		messagesChannel:         make(chan *interfaces.ConsensusRawMessage),
		acknowledgeBlockChannel: make(chan *blockWithProof),
		currentHeight:           0,
		config:                  config,
		logger:                  config.Logger,
		filter:                  filter,
		onCommitCallback:        onCommitCallback,
	}
}

func (lh *LeanHelix) Run(ctx context.Context) {
	lh.logger.Info("Run() starting infinite loop")
	for {
		select {
		case <-ctx.Done():
			lh.logger.Debug("LHFLOW Run Done")
			return

		case message := <-lh.messagesChannel:
			lh.filter.HandleConsensusRawMessage(ctx, message)

		case trigger := <-lh.config.ElectionTrigger.ElectionChannel():
			lh.logger.Debug("LHFLOW Run Election")
			if trigger == nil {
				lh.logger.Debug("LHFLOW Run Election, OMG trigger is nil!")
			}
			trigger(ctx)

		case blockWithProof := <-lh.acknowledgeBlockChannel:
			lh.logger.Debug("LHFLOW Run Update")
			prevHeight := blockheight.GetBlockHeight(blockWithProof.block)
			if prevHeight >= lh.currentHeight {
				lh.logger.Debug("Calling onNewConsensusRound() from Run() prevHeight=%d lh.currentHeight=%d", prevHeight, lh.currentHeight)
				lh.onNewConsensusRound(ctx, blockWithProof.block, blockWithProof.prevBlockProofBytes)
			}
		}
	}
}

func (lh *LeanHelix) UpdateState(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {
	height := blockheight.GetBlockHeight(prevBlock)
	lh.logger.Debug("LHFLOW UpdateState() ID=%s prevBlockHeight=%d", termincommittee.Str(lh.config.Membership.MyMemberId()), height)

	select {
	case <-ctx.Done():
		return

	case lh.acknowledgeBlockChannel <- &blockWithProof{prevBlock, prevBlockProofBytes}:
		return
	}

}

func (lh *LeanHelix) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, prevBlockProofBytes []byte) bool {
	lh.logger.Debug("ValidateBlockConsensus() ID=%s", termincommittee.Str(lh.config.Membership.MyMemberId()))
	if blockProofBytes == nil || len(blockProofBytes) == 0 || block == nil {
		return false
	}

	blockProof := protocol.BlockProofReader(blockProofBytes)
	blockRef := blockProof.BlockRef()
	if blockRef.MessageType() != protocol.LEAN_HELIX_COMMIT {
		return false
	}

	blockHeight := block.Height()
	if blockHeight != blockRef.BlockHeight() {
		return false
	}

	if !lh.config.BlockUtils.ValidateBlockCommitment(blockHeight, block, blockRef.BlockHash()) {
		return false
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
		if !proofsvalidator.VerifyBlockRefMessage(blockRef, sender, lh.config.KeyManager) {
			return false
		}

		memberId := sender.MemberId()
		if _, ok := set[storage.MemberIdStr(memberId)]; ok {
			return false
		}

		if !proofsvalidator.IsInMembers(committeeMembers, memberId) {
			return false
		}

		set[storage.MemberIdStr(memberId)] = true
		sendersCounter++
	}

	if sendersCounter < quorum.CalcQuorumSize(len(committeeMembers)) {
		return false
	}

	if len(blockProof.RandomSeedSignature()) == 0 || blockProof.RandomSeedSignature() == nil {
		return false
	}

	prevBlockProof := protocol.BlockProofReader(prevBlockProofBytes)
	if !randomseed.ValidateRandomSeed(lh.config.KeyManager, blockHeight, blockProof, prevBlockProof) {
		return false
	}

	return true
}

func (lh *LeanHelix) HandleConsensusRawMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {
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
