package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type LeanHelix struct {
	messagesChannel         chan ConsensusRawMessage
	acknowledgeBlockChannel chan Block
	currentHeight           primitives.BlockHeight
	config                  *Config
	logger                  Logger
	filter                  *ConsensusMessageFilter
	leanHelixTerm           *LeanHelixTerm
	commitSubscriptions     []func(block Block)
}

func (lh *LeanHelix) notifyCommitted(block Block) {
	lh.logger.Debug("%s LeanHelix.notifyCommitted(), %s", lh.config.Membership.MyMemberId().KeyForMap(), block)
	for _, subscription := range lh.commitSubscriptions {
		subscription(block)
	}
}

func (lh *LeanHelix) IsLeader() bool {
	return lh.leanHelixTerm != nil && lh.leanHelixTerm.IsLeader()
}

func (lh *LeanHelix) RegisterOnCommitted(cb func(block Block)) {
	lh.commitSubscriptions = append(lh.commitSubscriptions, cb)
}

func (lh *LeanHelix) GossipMessageReceived(ctx context.Context, msg ConsensusRawMessage) {
	lh.messagesChannel <- msg
}

func (lh *LeanHelix) ValidateBlockConsensus(block Block, blockProof *protocol.BlockProof, prevBlockProof *protocol.BlockProof) bool {
	// TODO: implement after 16-DEC-2018 - spec on lh-outline is incomplete!
	return true
}

func (lh *LeanHelix) Run(ctx context.Context) {
	for {
		if !lh.Tick(ctx) {
			return
		}
	}
}

func (lh *LeanHelix) Tick(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false

	case message := <-lh.messagesChannel:
		lh.filter.GossipMessageReceived(ctx, message)

	case trigger := <-lh.getElectionChannel():
		lh.logger.Info("Tick() election")
		trigger(ctx)

	case prevBlock := <-lh.acknowledgeBlockChannel:
		if prevBlock.Height() >= lh.currentHeight {
			lh.onNewConsensusRound(ctx, prevBlock)
		}
	}

	return true
}

func (lh *LeanHelix) UpdateConsensusRound(prevBlock Block) {
	lh.acknowledgeBlockChannel <- prevBlock
}

func (lh *LeanHelix) getElectionChannel() chan func(ctx context.Context) {
	if lh.leanHelixTerm == nil {
		return nil
	}
	return lh.leanHelixTerm.electionTrigger.ElectionChannel()
}

func (lh *LeanHelix) onCommit(ctx context.Context, block Block) {
	lh.notifyCommitted(block)
	lh.onNewConsensusRound(ctx, block)
}

func (lh *LeanHelix) onNewConsensusRound(ctx context.Context, prevBlock Block) {
	lh.currentHeight = prevBlock.Height() + 1
	lh.leanHelixTerm = NewLeanHelixTerm(ctx, lh.config, lh.onCommit, prevBlock)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
	lh.leanHelixTerm.StartTerm(ctx)
}

func NewLeanHelix(config *Config) *LeanHelix {
	if config.Logger == nil {
		config.Logger = NewSilentLogger()
	}
	filter := NewConsensusMessageFilter(config.Membership.MyMemberId(), config.Logger)
	return &LeanHelix{
		messagesChannel:         make(chan ConsensusRawMessage),
		acknowledgeBlockChannel: make(chan Block),
		currentHeight:           0,
		config:                  config,
		logger:                  config.Logger,
		filter:                  filter,
	}
}
