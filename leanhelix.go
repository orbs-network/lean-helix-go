package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type LeanHelix struct {
	messagesChannel         chan ConsensusRawMessage
	acknowledgeBlockChannel chan Block
	currentHeight           primitives.BlockHeight
	config                  *Config
	filter                  *ConsensusMessageFilter
	leanHelixTerm           *LeanHelixTerm
	commitSubscriptions     []func(block Block)
}

func (lh *LeanHelix) notifyCommitted(block Block) {
	for _, subscription := range lh.commitSubscriptions {
		subscription(block)
	}
}

func (lh *LeanHelix) RegisterOnCommitted(cb func(block Block)) {
	lh.commitSubscriptions = append(lh.commitSubscriptions, cb)
}

func (lh *LeanHelix) GossipMessageReceived(ctx context.Context, msg ConsensusRawMessage) {
	lh.messagesChannel <- msg
}

func (lh *LeanHelix) ValidateBlockConsensus(block Block, blockProof *BlockProof, prevBlockProof *BlockProof) bool {
	// TODO: implement
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
		trigger(ctx)

	case prevBlock := <-lh.acknowledgeBlockChannel:
		if prevBlock.Height() >= lh.currentHeight {
			lh.onNewConsensusRound(ctx, prevBlock.Height()+1)
		}
	}

	return true
}

func (lh *LeanHelix) AcknowledgeBlockConsensus(prevBlock Block) {
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
	lh.onNewConsensusRound(ctx, block.Height()+1)
}

func (lh *LeanHelix) onNewConsensusRound(ctx context.Context, height primitives.BlockHeight) {
	lh.currentHeight = height
	lh.leanHelixTerm = NewLeanHelixTerm(lh.config, lh.onCommit, lh.currentHeight)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
	lh.leanHelixTerm.StartTerm(ctx)
}

func NewLeanHelix(config *Config) *LeanHelix {
	filter := NewConsensusMessageFilter(config.KeyManager.MyPublicKey())
	return &LeanHelix{
		messagesChannel:         make(chan ConsensusRawMessage),
		acknowledgeBlockChannel: make(chan Block),
		currentHeight:           0,
		config:                  config,
		filter:                  filter,
	}
}
