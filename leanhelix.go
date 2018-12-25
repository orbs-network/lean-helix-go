package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"math"
)

type LeanHelix struct {
	messagesChannel         chan *ConsensusRawMessage
	acknowledgeBlockChannel chan Block
	currentHeight           primitives.BlockHeight
	config                  *Config
	logger                  Logger
	filter                  *ConsensusMessageFilter
	leanHelixTerm           *LeanHelixTerm
	onCommitCallback        OnCommitCallback
}

var GenesisBlock Block = nil

type OnCommitCallback func(ctx context.Context, block Block, blockProof []byte)

// ***********************************
// LeanHelix Constructor
// ***********************************
func NewLeanHelix(config *Config, onCommitCallback OnCommitCallback) *LeanHelix {
	if config.Logger == nil {
		config.Logger = NewSilentLogger()
	}

	config.Logger.Debug("NewLeanHelix() ID=%s", Str(config.Membership.MyMemberId()))
	filter := NewConsensusMessageFilter(config.Membership.MyMemberId(), config.Logger)
	return &LeanHelix{
		messagesChannel:         make(chan *ConsensusRawMessage),
		acknowledgeBlockChannel: make(chan Block),
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
		if !lh.Tick(ctx) {
			lh.logger.Info("Run() stopped infinite loop")
			return
		}
	}
}

func (lh *LeanHelix) UpdateState(prevBlock Block) {
	var height primitives.BlockHeight
	if prevBlock == nil {
		height = 0
	} else {
		height = prevBlock.Height()
	}
	lh.logger.Debug("UpdateState() ID=%s prevBlockHeight=%d", Str(lh.config.Membership.MyMemberId()), height)
	lh.acknowledgeBlockChannel <- prevBlock
}

func CalcQuorumSize(committeeMembersCount int) int {
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}

func (lh *LeanHelix) ValidateBlockConsensus(ctx context.Context, block Block, blockProofBytes []byte) bool {
	lh.logger.Debug("ValidateBlockConsensus() ID=%s", Str(lh.config.Membership.MyMemberId()))
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

	cSendersIterator := blockProof.NodesIterator()
	var sendersCounter = 0
	for {
		if !cSendersIterator.HasNext() {
			break
		}
		if !verifyBlockRefMessage(blockRef, cSendersIterator.NextNodes(), lh.config.KeyManager) {
			return false
		}
		sendersCounter++
	}

	committeeMembers := lh.config.Membership.RequestOrderedCommittee(ctx, blockHeight, 0, 9999)
	if sendersCounter < CalcQuorumSize(len(committeeMembers)) {
		return false
	}

	return true
}

func (lh *LeanHelix) HandleConsensusMessage(ctx context.Context, message *ConsensusRawMessage) {
	lh.logger.Debug("HandleConsensusRawMessage() ID=%s", Str(lh.config.Membership.MyMemberId()))
	select {
	case <-ctx.Done():
		return

	case lh.messagesChannel <- message:
	}
}

func (lh *LeanHelix) Tick(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false

	case message := <-lh.messagesChannel:
		lh.filter.HandleConsensusRawMessage(ctx, message)

	case trigger := <-lh.config.ElectionTrigger.ElectionChannel():
		trigger(ctx)

	case prevBlock := <-lh.acknowledgeBlockChannel:
		// TODO: a byzantine node can send the genesis block in sync can cause a mess
		var prevHeight primitives.BlockHeight
		if prevBlock == GenesisBlock {
			prevHeight = 0
		} else {
			prevHeight = prevBlock.Height()
		}
		if prevHeight >= lh.currentHeight {
			lh.onNewConsensusRound(ctx, prevBlock)
		}
	}

	return true
}

// ************************ Internal ***************************************

func (lh *LeanHelix) onCommit(ctx context.Context, block Block, blockProof []byte) {
	lh.logger.Debug("onCommit()")
	lh.onCommitCallback(ctx, block, blockProof)
	lh.onNewConsensusRound(ctx, block)
}

func (lh *LeanHelix) onNewConsensusRound(ctx context.Context, prevBlock Block) {
	if prevBlock == GenesisBlock {
		lh.currentHeight = 1
	} else {
		lh.currentHeight = primitives.BlockHeight(prevBlock.Height()) + 1
	}
	lh.leanHelixTerm = NewLeanHelixTerm(ctx, lh.config, lh.onCommit, prevBlock)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
}

func Str(memberId primitives.MemberId) string {
	return memberId.String()[:6]
}
