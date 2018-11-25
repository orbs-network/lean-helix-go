package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type ConsensusMessageFilter struct {
	myPublicKey     primitives.Ed25519PublicKey
	messagesChannel chan ConsensusRawMessage
	messageCache    map[primitives.BlockHeight][]ConsensusMessage
	logger          Logger
}

func NewConsensusMessageFilter(myPublicKey primitives.Ed25519PublicKey, logger Logger) *ConsensusMessageFilter {
	res := &ConsensusMessageFilter{
		myPublicKey:     myPublicKey,
		messagesChannel: make(chan ConsensusRawMessage, 0),
		messageCache:    make(map[primitives.BlockHeight][]ConsensusMessage),
		logger:          logger,
	}

	return res
}

func (f *ConsensusMessageFilter) WaitForMessage(ctx context.Context, blockHeight primitives.BlockHeight) (ConsensusMessage, error) {
	f.logger.Debug("WaitForMessage() start, blockHeight=%v", blockHeight)
	message := f.popFromCache(blockHeight)
	if message != nil {
		return message, nil
	}
	f.logger.Debug("WaitForMessage() popped")
	for {
		select {
		case <-ctx.Done():
			f.logger.Debug("WaitForMessage() Done()")
			return nil, ctx.Err()

		case rawMessage := <-f.messagesChannel:
			f.logger.Debug("WaitForMessage() got message")
			message = rawMessage.ToConsensusMessage()
			if f.isMyMessage(message) {
				f.logger.Debug("WaitForMessage() rejected: from me")
				continue
			}

			if message.BlockHeight() > blockHeight {
				f.logger.Debug("WaitForMessage() rejected: from future")
				f.pushToCache(message.BlockHeight(), message)
				continue
			}

			if message.BlockHeight() < blockHeight {
				f.logger.Debug("WaitForMessage() rejected: from past")
				continue
			}

			f.logger.Debug("WaitForMessage() accepted")
			return message, nil
		}
	}
}

// This method must run in a different goroutine than the consensus goroutine
func (f *ConsensusMessageFilter) OnGossipMessage(ctx context.Context, message ConsensusRawMessage) {
	f.messagesChannel <- message
}

func (f *ConsensusMessageFilter) isMyMessage(message ConsensusMessage) bool {
	return f.myPublicKey.Equal(message.SenderPublicKey())
}

func (f *ConsensusMessageFilter) clearCacheHistory(height primitives.BlockHeight) {
	for messageHeight := range f.messageCache {
		if messageHeight < height {
			delete(f.messageCache, messageHeight)
		}
	}
}

func (f *ConsensusMessageFilter) popFromCache(height primitives.BlockHeight) ConsensusMessage {
	f.clearCacheHistory(height)

	messages := f.messageCache[height]
	if messages == nil || len(messages) == 0 {
		return nil
	}

	message := messages[0]
	f.messageCache[height] = messages[1:]

	return message
}

func (f *ConsensusMessageFilter) pushToCache(height primitives.BlockHeight, message ConsensusMessage) {
	if f.messageCache[height] == nil {
		f.messageCache[height] = []ConsensusMessage{message}
	} else {
		f.messageCache[height] = append(f.messageCache[height], message)
	}
}
