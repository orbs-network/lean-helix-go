package rawmessagesfilter

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
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
	fmt.Printf("HandleConsensusRawMessage(): LHFILTER RECEIVED %s\n", message.MessageType())
	//f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER RECEIVED %s with H=%d V=%d sender=%s", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
	if f.isMyMessage(message) {
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER RECEIVED %s with H=%d V=%d sender=%s IGNORING message I sent", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
		return
	}

	if message.BlockHeight() < f.blockHeight {
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER RECEIVED %s with H=%d V=%d sender=%s IGNORING message from the past", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
		return
	}

	if message.InstanceId() != f.instanceId {
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER RECEIVED %s with H=%d V=%d sender=%s IGNORING message from different instanceID=%s because my instanceID==%s", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()), message.InstanceId(), f.instanceId)
		return
	}

	if message.BlockHeight() > f.blockHeight {
		f.pushToCache(message.BlockHeight(), message)
		f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER RECEIVED %s with H=%d V=%d sender=%s STORING message from future height", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
		return
	}
	f.logger.Debug(L.LC(f.blockHeight, 0, f.myMemberId), "LHFILTER RECEIVED %s with H=%d V=%d sender=%s OK PROCESSING", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
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
