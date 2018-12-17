package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type ConsensusMessageFilter struct {
	blockHeight         primitives.BlockHeight
	termMessagesHandler TermMessagesHandler
	myMemberId          primitives.MemberId
	messageCache        map[primitives.BlockHeight][]ConsensusMessage
	logger              Logger
}

func NewConsensusMessageFilter(myMemberId primitives.MemberId, logger Logger) *ConsensusMessageFilter {
	res := &ConsensusMessageFilter{
		myMemberId:   myMemberId,
		messageCache: make(map[primitives.BlockHeight][]ConsensusMessage),
		logger:       logger,
	}

	return res
}

func (f *ConsensusMessageFilter) GossipMessageReceived(ctx context.Context, rawMessage *ConsensusRawMessage) {
	message := ToConsensusMessage(rawMessage)
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

func (f *ConsensusMessageFilter) isMyMessage(message ConsensusMessage) bool {
	return f.myMemberId.Equal(message.SenderMemberId())
}

func (f *ConsensusMessageFilter) clearCacheHistory(height primitives.BlockHeight) {
	for messageHeight := range f.messageCache {
		if messageHeight < height {
			delete(f.messageCache, messageHeight)
		}
	}
}

func (f *ConsensusMessageFilter) pushToCache(height primitives.BlockHeight, message ConsensusMessage) {
	if f.messageCache[height] == nil {
		f.messageCache[height] = []ConsensusMessage{message}
	} else {
		f.messageCache[height] = append(f.messageCache[height], message)
	}
}

func (f *ConsensusMessageFilter) processGossipMessage(ctx context.Context, message ConsensusMessage) {
	if f.termMessagesHandler == nil {
		return
	}

	switch message := message.(type) {
	case *PreprepareMessage:
		f.termMessagesHandler.HandleLeanHelixPrePrepare(ctx, message)
	case *PrepareMessage:
		f.termMessagesHandler.HandleLeanHelixPrepare(ctx, message)
	case *CommitMessage:
		f.termMessagesHandler.HandleLeanHelixCommit(ctx, message)
	case *ViewChangeMessage:
		f.termMessagesHandler.HandleLeanHelixViewChange(ctx, message)
	case *NewViewMessage:
		f.termMessagesHandler.HandleLeanHelixNewView(ctx, message)
	default:
		panic(fmt.Sprintf("unknown message type: %T", message))
	}
}

func (f *ConsensusMessageFilter) consumeCacheMessages(ctx context.Context, blockHeight primitives.BlockHeight) {
	f.clearCacheHistory(blockHeight)

	messages := f.messageCache[blockHeight]
	for _, message := range messages {
		f.processGossipMessage(ctx, message)
	}
	delete(f.messageCache, blockHeight)
}

func (f *ConsensusMessageFilter) SetBlockHeight(ctx context.Context, blockHeight primitives.BlockHeight, termMessagesHandler TermMessagesHandler) {
	f.termMessagesHandler = termMessagesHandler
	f.blockHeight = blockHeight
	f.consumeCacheMessages(ctx, blockHeight)
}
