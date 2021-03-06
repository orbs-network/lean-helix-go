// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

const BLOCK_HEIGHT_DONT_CARE = math.MaxUint64
const VIEW_DONT_CARE = math.MaxUint64

type SubscriptionValue struct {
	cb func(ctx context.Context, message *interfaces.ConsensusRawMessage)
}

type outgoingMessage struct {
	target  primitives.MemberId
	message *interfaces.ConsensusRawMessage
}

type messageProps struct {
	messageType protocol.MessageType
	height      primitives.BlockHeight
	view        primitives.View
	sender      primitives.MemberId
	receiver    primitives.MemberId
}

type CommunicationMock struct {
	memberId  primitives.MemberId
	discovery *Discovery
	logger    interfaces.Logger

	outgoingLock        sync.RWMutex
	outgoingChannelsMap map[string]chan *outgoingMessage

	subscriptionsLock   sync.Mutex
	nextSubscriptionKey int
	subscriptions       map[int]*SubscriptionValue

	muOut                      sync.RWMutex
	outgoingWhitelistMemberIds []primitives.MemberId

	muIn                       sync.RWMutex
	incomingWhiteListMemberIds []primitives.MemberId

	statsSentMessagesMutex sync.RWMutex
	statsSentMessages      []*interfaces.ConsensusRawMessage
	maxDelayDuration       time.Duration
	messagesHistoryLock    sync.Mutex
	messagesHistory        []*messageProps
}

func NewCommunication(memberId primitives.MemberId, discovery *Discovery, log interfaces.Logger) *CommunicationMock {

	if log == nil {
		log = logger.NewConsoleLogger("XXX")
	}

	return &CommunicationMock{
		memberId:                   memberId,
		discovery:                  discovery,
		logger:                     log,
		outgoingChannelsMap:        make(map[string]chan *outgoingMessage),
		nextSubscriptionKey:        0,
		subscriptions:              make(map[int]*SubscriptionValue),
		outgoingWhitelistMemberIds: nil,
		incomingWhiteListMemberIds: nil,
		statsSentMessages:          []*interfaces.ConsensusRawMessage{},
		maxDelayDuration:           time.Duration(0),
		messagesHistory:            []*messageProps{},
	}
}

func (g *CommunicationMock) SendConsensusMessage(ctx context.Context, targets []primitives.MemberId, message *interfaces.ConsensusRawMessage) error {
	g.statsSentMessagesMutex.Lock()
	defer g.statsSentMessagesMutex.Unlock()

	g.statsSentMessages = append(g.statsSentMessages, message)
	for _, target := range targets {
		channel := g.ReturnOutgoingChannelByTarget(target)
		select {
		default: // never block. ignore message if buffer is full
		case <-ctx.Done():
			return errors.Errorf("ID=%s context canceled for outgoing channel of %v", g.memberId, target)
		case channel <- &outgoingMessage{target, message}:
			msg := interfaces.ToConsensusMessage(message)
			g.logger.Debug("COMM: ID=%s Sent message %s to %s H=%d V=%d", msg.SenderMemberId(), msg.MessageType(), target,
				msg.BlockHeight(), msg.View())
			continue
		}
	}
	return nil
}

func (g *CommunicationMock) messageSenderLoop(ctx context.Context, channel chan *outgoingMessage) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("messageSenderLoop() PANIC: ", e)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case messageData := <-channel:
			g.SendToNode(ctx, messageData.target, messageData.message)
		}

	}
}

func (g *CommunicationMock) SetMessagesMaxDelay(duration time.Duration) {
	g.maxDelayDuration = duration
}

func (g *CommunicationMock) OnIncomingMessage(ctx context.Context, rawMessage *interfaces.ConsensusRawMessage) {
	g.subscriptionsLock.Lock()
	defer g.subscriptionsLock.Unlock()

	if g.bannedSender(rawMessage) {
		return
	}

	count := 0
	for _, s := range g.subscriptions {
		if g.maxDelayDuration > 0 {
			time.Sleep(time.Duration(rand.Int63n(int64(g.maxDelayDuration))))
		}
		count++
		s.cb(ctx, rawMessage)
	}
	if count == 0 {
		parsedMessage := interfaces.ToConsensusMessage(rawMessage)
		g.logger.Error("failed delivery for %s - no subscriber found", parsedMessage.MessageType())
	}
}

func (g *CommunicationMock) addToHistory(rawMessage *interfaces.ConsensusRawMessage, receiver primitives.MemberId) {
	msg := interfaces.ToConsensusMessage(rawMessage)
	msgProps := &messageProps{
		messageType: msg.MessageType(),
		height:      msg.BlockHeight(),
		view:        msg.View(),
		sender:      msg.SenderMemberId(),
		receiver:    receiver,
	}
	//fmt.Printf("ID=%s addToHistory(): H=%d V=%d TYPE=%s sender=%s receiver=%s\n",
	//	g.memberId, msgProps.height, msgProps.view, msgProps.messageType, msgProps.sender, msgProps.receiver)

	g.messagesHistoryLock.Lock()
	defer g.messagesHistoryLock.Unlock()
	g.messagesHistory = append(g.messagesHistory, msgProps)
}

func (g *CommunicationMock) SendToNode(ctx context.Context, receiverMemberId primitives.MemberId, consensusRawMessage *interfaces.ConsensusRawMessage) {
	if g.bannedReceiver(receiverMemberId) {
		return
	}

	if receiverCommunication := g.discovery.GetCommunicationById(receiverMemberId); receiverCommunication != nil {
		g.addToHistory(consensusRawMessage, receiverMemberId)
		receiverCommunication.OnIncomingMessage(ctx, consensusRawMessage)
	} else {
		return
	}
	return
}

func (g *CommunicationMock) bannedReceiver(receiverMemberId primitives.MemberId) bool {
	g.muOut.RLock()
	defer g.muOut.RUnlock()
	if g.outgoingWhitelistMemberIds == nil {
		return false
	}

	for _, currentId := range g.outgoingWhitelistMemberIds {
		if currentId.Equal(receiverMemberId) {
			return false
		}
	}
	return true
}

func (g *CommunicationMock) bannedSender(rawMessage *interfaces.ConsensusRawMessage) bool {
	g.muIn.RLock()
	defer g.muIn.RUnlock()

	if g.incomingWhiteListMemberIds == nil {
		return false
	}

	senderMemberId := interfaces.ToConsensusMessage(rawMessage).SenderMemberId()
	for _, currentId := range g.incomingWhiteListMemberIds {
		if currentId.Equal(senderMemberId) {
			return false
		}
	}
	return true
}

func (g *CommunicationMock) ReturnOutgoingChannelByTarget(target primitives.MemberId) chan *outgoingMessage {
	g.outgoingLock.RLock()
	defer g.outgoingLock.RUnlock()

	return g.outgoingChannelsMap[target.String()]
}

func (g *CommunicationMock) ReturnAndMaybeCreateOutgoingChannelByTarget(ctx context.Context, target primitives.MemberId) chan *outgoingMessage {
	g.outgoingLock.Lock()
	defer g.outgoingLock.Unlock()

	channel := g.outgoingChannelsMap[target.String()]
	if channel == nil {
		channel = make(chan *outgoingMessage, 1000)
		g.outgoingChannelsMap[target.String()] = channel
		go g.messageSenderLoop(ctx, channel)
	}

	return channel
}

func (g *CommunicationMock) RegisterIncomingMessageHandler(cb func(ctx context.Context, message *interfaces.ConsensusRawMessage)) {
	g.subscriptionsLock.Lock()
	defer g.subscriptionsLock.Unlock()

	g.nextSubscriptionKey++
	g.subscriptions[g.nextSubscriptionKey] = &SubscriptionValue{cb}
}

func (g *CommunicationMock) inOutgoingWhitelist(memberId primitives.MemberId) bool {
	g.muOut.RLock()
	defer g.muOut.RUnlock()

	for _, currentId := range g.outgoingWhitelistMemberIds {
		if currentId.Equal(memberId) {
			return true
		}
	}
	return false
}

func (g *CommunicationMock) DisableOutgoingCommunication() {
	g.muOut.Lock()
	defer g.muOut.Unlock()
	g.outgoingWhitelistMemberIds = []primitives.MemberId{}
}

func (g *CommunicationMock) EnableOutgoingCommunication() {
	g.muOut.Lock()
	defer g.muOut.Unlock()
	g.outgoingWhitelistMemberIds = nil
}

func (g *CommunicationMock) SetOutgoingWhitelist(outgoingWhitelist []primitives.MemberId) {
	g.muOut.Lock()
	defer g.muOut.Unlock()
	if len(outgoingWhitelist) == 0 {
		panic("Instead of setting nil, use EnableOutgoingCommunication(). Instead of setting empty array use DisableOutgoingCommunication()")
	}
	g.outgoingWhitelistMemberIds = outgoingWhitelist
}

func (g *CommunicationMock) DisableIncomingCommunication() {
	g.muIn.Lock()
	defer g.muIn.Unlock()

	g.incomingWhiteListMemberIds = []primitives.MemberId{}
}

func (g *CommunicationMock) EnableIncomingCommunication() {
	g.muIn.Lock()
	defer g.muIn.Unlock()

	g.incomingWhiteListMemberIds = nil
}

func (g *CommunicationMock) SetIncomingWhitelist(incomingWhitelist []primitives.MemberId) {
	g.muIn.Lock()
	defer g.muIn.Unlock()

	if len(incomingWhitelist) == 0 {
		panic("Instead of setting nil, use EnableIncomingCommunication(). Instead of setting empty array use DisableIncomingCommunication()")
	}
	g.incomingWhiteListMemberIds = incomingWhitelist
}

// TODO REMOVE THIS REDUNDANT METHOD
func (g *CommunicationMock) CountSentMessages(messageType protocol.MessageType) int {
	g.statsSentMessagesMutex.RLock()
	defer g.statsSentMessagesMutex.RUnlock()

	res := 0
	for _, msg := range g.statsSentMessages {
		if interfaces.ToConsensusMessage(msg).MessageType() == messageType {
			res++
		}
	}
	return res
}

// TODO Refactor this, maybe get the data from messagesHistory instead of statsSentMessages
func (g *CommunicationMock) GetSentMessages(messageType protocol.MessageType) []*interfaces.ConsensusRawMessage {
	g.statsSentMessagesMutex.RLock()
	defer g.statsSentMessagesMutex.RUnlock()

	var res []*interfaces.ConsensusRawMessage
	for _, msg := range g.statsSentMessages {
		if interfaces.ToConsensusMessage(msg).MessageType() == messageType {
			res = append(res, msg)
		}
	}
	return res
}

func (g *CommunicationMock) CountMessagesSent(messageType protocol.MessageType, height primitives.BlockHeight, view primitives.View, target primitives.MemberId) int {

	g.messagesHistoryLock.Lock()
	defer g.messagesHistoryLock.Unlock()

	var counter int
	for _, msg := range g.messagesHistory {
		if messageType != msg.messageType {
			continue
		}
		if !g.memberId.Equal(msg.sender) {
			continue
		}
		if target != nil && !target.Equal(msg.receiver) {
			continue
		}
		if height != BLOCK_HEIGHT_DONT_CARE && height != msg.height {
			continue
		}
		if view != VIEW_DONT_CARE && view != msg.view {
			continue
		}
		counter++
	}
	return counter

}
