// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"math/rand"
	"time"
)

type SubscriptionValue struct {
	cb func(ctx context.Context, message *interfaces.ConsensusRawMessage)
}

type outgoingMessage struct {
	target  primitives.MemberId
	message *interfaces.ConsensusRawMessage
}

type CommunicationMock struct {
	discovery                  *Discovery
	outgoingChannelsMap        map[string]chan *outgoingMessage
	totalSubscriptions         int
	subscriptions              map[int]*SubscriptionValue
	outgoingWhitelist          []primitives.MemberId
	incomingWhiteListMemberIds []primitives.MemberId
	statsSentMessages          []*interfaces.ConsensusRawMessage
	maxDelayDuration           time.Duration
}

func NewCommunication(discovery *Discovery) *CommunicationMock {
	return &CommunicationMock{
		discovery:                  discovery,
		outgoingChannelsMap:        make(map[string]chan *outgoingMessage),
		totalSubscriptions:         0,
		subscriptions:              make(map[int]*SubscriptionValue),
		outgoingWhitelist:          nil,
		incomingWhiteListMemberIds: nil,
		statsSentMessages:          []*interfaces.ConsensusRawMessage{},
		maxDelayDuration:           time.Duration(0),
	}
}

func (g *CommunicationMock) SetMessagesMaxDelay(duration time.Duration) {
	g.maxDelayDuration = duration
}

func (g *CommunicationMock) messageSenderLoop(ctx context.Context, channel chan *outgoingMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		case messageData := <-channel:
			g.SendToNode(ctx, messageData.target, messageData.message)
		}

	}
}

func (g *CommunicationMock) getOutgoingChannelByTarget(ctx context.Context, target primitives.MemberId) chan *outgoingMessage {
	channel := g.outgoingChannelsMap[target.String()]
	if channel == nil {
		channel = make(chan *outgoingMessage, 100)
		g.outgoingChannelsMap[target.String()] = channel
		go g.messageSenderLoop(ctx, channel)
	}

	return channel
}

func (g *CommunicationMock) SendConsensusMessage(ctx context.Context, targets []primitives.MemberId, message *interfaces.ConsensusRawMessage) {
	g.statsSentMessages = append(g.statsSentMessages, message)
	for _, target := range targets {
		channel := g.getOutgoingChannelByTarget(ctx, target)
		select {
		case <-ctx.Done():
			return
		case channel <- &outgoingMessage{target, message}:
		}
	}
}

func (g *CommunicationMock) RegisterOnMessage(cb func(ctx context.Context, message *interfaces.ConsensusRawMessage)) int {
	g.totalSubscriptions++
	g.subscriptions[g.totalSubscriptions] = &SubscriptionValue{cb}
	return g.totalSubscriptions
}

func (g *CommunicationMock) UnregisterOnMessage(subscriptionToken int) {
	delete(g.subscriptions, subscriptionToken)
}

func (g *CommunicationMock) OnRemoteMessage(ctx context.Context, rawMessage *interfaces.ConsensusRawMessage) {
	for _, s := range g.subscriptions {
		if g.incomingWhiteListMemberIds != nil {
			senderMemberId := interfaces.ToConsensusMessage(rawMessage).SenderMemberId()
			if !g.inIncomingWhitelist(senderMemberId) {
				continue
			}
		}

		//go func() {
		if g.maxDelayDuration > 0 {
			time.Sleep(time.Duration(rand.Int63n(int64(g.maxDelayDuration))))
		}
		s.cb(ctx, rawMessage)
		//}()
	}
}

func (g *CommunicationMock) inIncomingWhitelist(memberId primitives.MemberId) bool {
	for _, currentId := range g.incomingWhiteListMemberIds {
		if currentId.Equal(memberId) {
			return true
		}
	}
	return false
}

func (g *CommunicationMock) inOutgoingWhitelist(memberId primitives.MemberId) bool {
	for _, currentId := range g.outgoingWhitelist {
		if currentId.Equal(memberId) {
			return true
		}
	}
	return false
}

func (g *CommunicationMock) DisableOutgoing() {
	g.SetOutgoingWhitelist([]primitives.MemberId{})
}

func (g *CommunicationMock) EnableOutgoing() {
	g.ClearOutgoingWhitelist()
}

func (g *CommunicationMock) SetOutgoingWhitelist(outgoingWhitelist []primitives.MemberId) {
	g.outgoingWhitelist = outgoingWhitelist
}

func (g *CommunicationMock) ClearOutgoingWhitelist() {
	g.SetOutgoingWhitelist(nil)
}

func (g *CommunicationMock) SetIncomingWhitelist(incomingWhitelist []primitives.MemberId) {
	g.incomingWhiteListMemberIds = incomingWhitelist
}

func (g *CommunicationMock) ClearIncomingWhitelist() {
	g.SetIncomingWhitelist(nil)
}

func (g *CommunicationMock) SendToNode(ctx context.Context, targetMemberId primitives.MemberId, consensusRawMessage *interfaces.ConsensusRawMessage) {
	if g.outgoingWhitelist != nil {
		if !g.inOutgoingWhitelist(targetMemberId) {
			return
		}
	}

	if targetCommunication := g.discovery.GetCommunicationById(targetMemberId); targetCommunication != nil {
		targetCommunication.OnRemoteMessage(ctx, consensusRawMessage)
	}
	return
}

func (g *CommunicationMock) CountSentMessages(messageType protocol.MessageType) int {
	res := 0
	for _, msg := range g.statsSentMessages {
		if interfaces.ToConsensusMessage(msg).MessageType() == messageType {
			res++
		}
	}
	return res
}

func (g *CommunicationMock) GetSentMessages(messageType protocol.MessageType) []*interfaces.ConsensusRawMessage {
	var res []*interfaces.ConsensusRawMessage
	for _, msg := range g.statsSentMessages {
		if interfaces.ToConsensusMessage(msg).MessageType() == messageType {
			res = append(res, msg)
		}
	}
	return res
}
