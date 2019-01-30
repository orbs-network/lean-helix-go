package rawmessagesfilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type RawMessageFilter struct {
	instanceId               primitives.InstanceId
	blockHeight              primitives.BlockHeight
	consensusMessagesHandler ConsensusMessagesHandler
	myMemberId               primitives.MemberId
	messageCache             map[primitives.BlockHeight][]interfaces.ConsensusMessage
	logger                   L.LHLogger
}

func NewConsensusMessageFilter(instanceId primitives.InstanceId, myMemberId primitives.MemberId, logger L.LHLogger) *RawMessageFilter {
	res := &RawMessageFilter{
		instanceId:   instanceId,
		myMemberId:   myMemberId,
		messageCache: make(map[primitives.BlockHeight][]interfaces.ConsensusMessage),
		logger:       logger,
	}

	return res
}

func (f *RawMessageFilter) HandleConsensusRawMessage(ctx context.Context, rawMessage *interfaces.ConsensusRawMessage) {
	message := interfaces.ToConsensusMessage(rawMessage)
	if f.isMyMessage(message) {
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER ignoring message I sent")
		return
	}

	if message.BlockHeight() < f.blockHeight {
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER ignoring message with height=%d because filter height=%d", message.BlockHeight(), f.blockHeight)
		return
	}

	if message.InstanceId() != f.instanceId {
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER ignoring message with instanceID=%s because my instanceID==%s", message.InstanceId(), f.instanceId)
		return
	}

	if message.BlockHeight() > f.blockHeight {
		f.pushToCache(message.BlockHeight(), message)
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER pushed to cache message from future block height=%d where filter height=%d", message.BlockHeight(), f.blockHeight)
		return
	}

	f.processConsensusMessage(ctx, message)
}

func (f *RawMessageFilter) isMyMessage(message interfaces.ConsensusMessage) bool {
	return f.myMemberId.Equal(message.SenderMemberId())
}

func (f *RawMessageFilter) clearCacheHistory(height primitives.BlockHeight) {
	for messageHeight := range f.messageCache {
		if messageHeight < height {
			delete(f.messageCache, messageHeight)
		}
	}
}

func (f *RawMessageFilter) pushToCache(height primitives.BlockHeight, message interfaces.ConsensusMessage) {
	if f.messageCache[height] == nil {
		f.messageCache[height] = []interfaces.ConsensusMessage{message}
	} else {
		f.messageCache[height] = append(f.messageCache[height], message)
	}
}

func (f *RawMessageFilter) processConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	if f.consensusMessagesHandler == nil {
		return
	}

	f.consensusMessagesHandler.HandleConsensusMessage(ctx, message)
}

func (f *RawMessageFilter) consumeCacheMessages(ctx context.Context, blockHeight primitives.BlockHeight) {
	f.clearCacheHistory(blockHeight)

	messages := f.messageCache[blockHeight]
	if len(messages) > 0 {
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER consuming %d messages from height=%d", len(messages), blockHeight)
	}
	for _, message := range messages {
		f.processConsensusMessage(ctx, message)
	}
	delete(f.messageCache, blockHeight)
}

func (f *RawMessageFilter) SetBlockHeight(ctx context.Context, blockHeight primitives.BlockHeight, consensusMessagesHandler ConsensusMessagesHandler) {
	f.consensusMessagesHandler = consensusMessagesHandler
	f.blockHeight = blockHeight
	f.consumeCacheMessages(ctx, blockHeight)
}
