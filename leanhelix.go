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

func GetBlockHeight(prevBlock Block) primitives.BlockHeight {
	if prevBlock == GenesisBlock {
		return 0
	} else {
		return prevBlock.Height()
	}
}

func CalcQuorumSize(committeeMembersCount int) int {
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}

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

func (lh *LeanHelix) UpdateState(prevBlock Block, blockProofBytes []byte) {
	var height primitives.BlockHeight
	if prevBlock == nil {
		height = 0
	} else {
		height = prevBlock.Height()
	}
	lh.logger.Debug("UpdateState() ID=%s prevBlockHeight=%d", Str(lh.config.Membership.MyMemberId()), height)
	lh.acknowledgeBlockChannel <- prevBlock
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

	committeeMembers := lh.config.Membership.RequestOrderedCommittee(ctx, blockHeight, 0)

	sendersIterator := blockProof.NodesIterator()
	set := make(map[MemberIdStr]bool)
	var sendersCounter = 0
	for {
		if !sendersIterator.HasNext() {
			break
		}

		sender := sendersIterator.NextNodes()
		if !verifyBlockRefMessage(blockRef, sender, lh.config.KeyManager) {
			return false
		}

		memberId := sender.MemberId()
		if _, ok := set[MemberIdStr(memberId)]; ok {
			return false
		}

		if !isInMembers(committeeMembers, memberId) {
			return false
		}

		set[MemberIdStr(memberId)] = true
		sendersCounter++
	}

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
		prevHeight := GetBlockHeight(prevBlock)
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
	lh.currentHeight = GetBlockHeight(prevBlock) + 1
	lh.leanHelixTerm = NewLeanHelixTerm(ctx, lh.config, lh.onCommit, prevBlock)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
}

func Str(memberId primitives.MemberId) string {
	return memberId.String()[:6]
}
