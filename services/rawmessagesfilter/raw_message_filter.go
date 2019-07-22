// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package rawmessagesfilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"math"
)

type RawMessageFilter struct {
	instanceId               primitives.InstanceId
	state                    state.State
	consensusMessagesHandler ConsensusMessagesHandler
	myMemberId               primitives.MemberId
	futureCache              map[primitives.BlockHeight][]interfaces.ConsensusMessage
	logger                   L.LHLogger
	latestFutureBlockHeight  primitives.BlockHeight // needed for limiting future cache to 1 term (potential memory leak)
}

func NewConsensusMessageFilter(instanceId primitives.InstanceId, myMemberId primitives.MemberId, logger L.LHLogger, state state.State) *RawMessageFilter {
	res := &RawMessageFilter{
		instanceId:  instanceId,
		myMemberId:  myMemberId,
		futureCache: make(map[primitives.BlockHeight][]interfaces.ConsensusMessage),
		logger:      logger,
		state:       state,
	}

	return res
}

func (f *RawMessageFilter) HandleConsensusRawMessage(ctx context.Context, rawMessage *interfaces.ConsensusRawMessage) {
	message := interfaces.ToConsensusMessage(rawMessage)
	if f.isMyMessage(message) {
		f.logger.Debug(L.LC(f.state.Height(), math.MaxUint64, f.myMemberId), "LHFILTER IGNORING RECEIVED %s with H=%d V=%d sender=%s IGNORING message I sent", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
		return
	}

	if message.BlockHeight() < f.state.Height() {
		f.logger.Debug(L.LC(f.state.Height(), math.MaxUint64, f.myMemberId), "LHFILTER IGNORING RECEIVED %s with H=%d V=%d sender=%s IGNORING message from the past", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
		return
	}

	if message.InstanceId() != f.instanceId {
		f.logger.Debug(L.LC(f.state.Height(), math.MaxUint64, f.myMemberId), "LHFILTER IGNORING RECEIVED %s with H=%d V=%d sender=%s IGNORING message from different instanceID=%s because my instanceID==%s", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()), message.InstanceId(), f.instanceId)
		return
	}

	if message.BlockHeight() > f.state.Height() {
		f.pushToCache(message.BlockHeight(), message)
		f.logger.Debug(L.LC(f.state.Height(), math.MaxUint64, f.myMemberId), "LHFILTER STORING RECEIVED %s with H=%d V=%d sender=%s STORING message from future height", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
		return
	}
	f.logger.Debug(L.LC(f.state.Height(), math.MaxUint64, f.myMemberId), "LHFILTER RECEIVED %s with H=%d V=%d sender=%s OK PROCESSING", message.MessageType(), message.BlockHeight(), message.View(), termincommittee.Str(message.SenderMemberId()))
	f.processConsensusMessage(ctx, message)
}

func (f *RawMessageFilter) isMyMessage(message interfaces.ConsensusMessage) bool {
	return f.myMemberId.Equal(message.SenderMemberId())
}

func (f *RawMessageFilter) clearCacheEarlierThan(height primitives.BlockHeight) {
	for messageHeight := range f.futureCache {
		if messageHeight < height {
			delete(f.futureCache, messageHeight)
		}
	}
}

func (f *RawMessageFilter) pushToCache(height primitives.BlockHeight, message interfaces.ConsensusMessage) {
	// limit future cache to 1 term - potential increased memory hogging linear to distance in block height between this node and currentBlockHeight (the "front")
	// TODO This requires testing https://github.com/orbs-network/lean-helix-go/issues/35
	if height < f.latestFutureBlockHeight {
		return
	}
	if height > f.latestFutureBlockHeight {
		f.clearCacheEarlierThan(height)
		f.latestFutureBlockHeight = height
	}

	// add to future cache
	if f.futureCache[height] == nil {
		f.futureCache[height] = []interfaces.ConsensusMessage{message}
	} else {
		f.futureCache[height] = append(f.futureCache[height], message)
	}
}

func (f *RawMessageFilter) processConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	if f.consensusMessagesHandler == nil {
		return
	}

	f.consensusMessagesHandler.HandleConsensusMessage(ctx, message)
}

func (f *RawMessageFilter) ConsumeCacheMessages(ctx context.Context, consensusMessagesHandler ConsensusMessagesHandler) {
	height := f.state.Height()
	f.consensusMessagesHandler = consensusMessagesHandler
	f.clearCacheEarlierThan(height)

	messages := f.futureCache[height]
	if len(messages) > 0 {
		f.logger.Debug(L.LC(f.state.Height(), math.MaxUint64, f.myMemberId), "LHFILTER consuming %d messages from height=%d", len(messages), height)
	}
	for _, message := range messages {
		f.processConsensusMessage(ctx, message)
	}
	delete(f.futureCache, height)
}
