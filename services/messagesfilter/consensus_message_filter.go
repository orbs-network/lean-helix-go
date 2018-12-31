package messagesfilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type ConsensusMessageFilter struct {
	blockHeight              primitives.BlockHeight
	consensusMessagesHandler ConsensusMessagesHandler
	myMemberId               primitives.MemberId
	messageCache             map[primitives.BlockHeight][]interfaces.ConsensusMessage
	logger                   interfaces.Logger
}

func NewConsensusMessageFilter(myMemberId primitives.MemberId, logger interfaces.Logger) *ConsensusMessageFilter {
	res := &ConsensusMessageFilter{
		myMemberId:   myMemberId,
		messageCache: make(map[primitives.BlockHeight][]interfaces.ConsensusMessage),
		logger:       logger,
	}

	return res
}

func (f *ConsensusMessageFilter) HandleConsensusRawMessage(ctx context.Context, rawMessage *interfaces.ConsensusRawMessage) {
	message := interfaces.ToConsensusMessage(rawMessage)
	if f.isMyMessage(message) {
		return
	}

	if message.BlockHeight() < f.blockHeight {
		return
	}

	if message.BlockHeight() > f.blockHeight {
		f.pushToCache(message.BlockHeight(), message)
		return
	}

	f.processGossipMessage(ctx, message)
}

func (f *ConsensusMessageFilter) isMyMessage(message interfaces.ConsensusMessage) bool {
	return f.myMemberId.Equal(message.SenderMemberId())
}

func (f *ConsensusMessageFilter) clearCacheHistory(height primitives.BlockHeight) {
	for messageHeight := range f.messageCache {
		if messageHeight < height {
			delete(f.messageCache, messageHeight)
		}
	}
}

func (f *ConsensusMessageFilter) pushToCache(height primitives.BlockHeight, message interfaces.ConsensusMessage) {
	if f.messageCache[height] == nil {
		f.messageCache[height] = []interfaces.ConsensusMessage{message}
	} else {
		f.messageCache[height] = append(f.messageCache[height], message)
	}
}

func (f *ConsensusMessageFilter) processGossipMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	if f.consensusMessagesHandler == nil {
		return
	}

	f.consensusMessagesHandler.HandleConsensusMessage(ctx, message)
}

func (f *ConsensusMessageFilter) consumeCacheMessages(ctx context.Context, blockHeight primitives.BlockHeight) {
	f.clearCacheHistory(blockHeight)

	messages := f.messageCache[blockHeight]
	for _, message := range messages {
		f.processGossipMessage(ctx, message)
	}
	delete(f.messageCache, blockHeight)
}

func (f *ConsensusMessageFilter) SetBlockHeight(ctx context.Context, blockHeight primitives.BlockHeight, consensusMessagesHandler ConsensusMessagesHandler) {
	f.consensusMessagesHandler = consensusMessagesHandler
	f.blockHeight = blockHeight
	f.consumeCacheMessages(ctx, blockHeight)
}
