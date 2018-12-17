package gossip

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type SubscriptionValue struct {
	cb func(ctx context.Context, message *leanhelix.ConsensusRawMessage)
}

type outgoingMessage struct {
	target  primitives.MemberId
	message *leanhelix.ConsensusRawMessage
}

type Gossip struct {
	discovery                  *Discovery
	outgoingChannelsMap        map[string]chan *outgoingMessage
	totalSubscriptions         int
	subscriptions              map[int]*SubscriptionValue
	outgoingWhitelist          []primitives.MemberId
	incomingWhiteListMemberIds []primitives.MemberId
	statsSentMessages          []*leanhelix.ConsensusRawMessage
}

func NewGossip(discovery *Discovery) *Gossip {
	return &Gossip{
		discovery:                  discovery,
		outgoingChannelsMap:        make(map[string]chan *outgoingMessage),
		totalSubscriptions:         0,
		subscriptions:              make(map[int]*SubscriptionValue),
		outgoingWhitelist:          nil,
		incomingWhiteListMemberIds: nil,
		statsSentMessages:          []*leanhelix.ConsensusRawMessage{},
	}
}

func (g *Gossip) messageSenderLoop(ctx context.Context, channel chan *outgoingMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		case messageData := <-channel:
			g.SendToNode(ctx, messageData.target, messageData.message)
		}

	}
}

func (g *Gossip) getOutgoingChannelByTarget(ctx context.Context, target primitives.MemberId) chan *outgoingMessage {
	channel := g.outgoingChannelsMap[target.String()]
	if channel == nil {
		channel = make(chan *outgoingMessage, 100)
		g.outgoingChannelsMap[target.String()] = channel
		go g.messageSenderLoop(ctx, channel)
	}

	return channel
}

func (g *Gossip) SendConsensusMessage(ctx context.Context, targets []primitives.MemberId, message *leanhelix.ConsensusRawMessage) {
	g.statsSentMessages = append(g.statsSentMessages, message)
	for _, target := range targets {
		channel := g.getOutgoingChannelByTarget(ctx, target)
		channel <- &outgoingMessage{target, message}
	}
}

func (g *Gossip) RegisterOnMessage(cb func(ctx context.Context, message *leanhelix.ConsensusRawMessage)) int {
	g.totalSubscriptions++
	g.subscriptions[g.totalSubscriptions] = &SubscriptionValue{cb}
	return g.totalSubscriptions
}

func (g *Gossip) UnregisterOnMessage(subscriptionToken int) {
	delete(g.subscriptions, subscriptionToken)
}

func (g *Gossip) OnRemoteMessage(ctx context.Context, rawMessage *leanhelix.ConsensusRawMessage) {
	for _, s := range g.subscriptions {
		if g.incomingWhiteListMemberIds != nil {
			senderMemberId := leanhelix.ToConsensusMessage(rawMessage).SenderMemberId()
			if !g.inIncomingWhitelist(senderMemberId) {
				continue
			}
		}
		s.cb(ctx, rawMessage)
	}
}

func (g *Gossip) inIncomingWhitelist(memberId primitives.MemberId) bool {
	for _, currentId := range g.incomingWhiteListMemberIds {
		if currentId.Equal(memberId) {
			return true
		}
	}
	return false
}

func (g *Gossip) inOutgoingWhitelist(memberId primitives.MemberId) bool {
	for _, currentId := range g.outgoingWhitelist {
		if currentId.Equal(memberId) {
			return true
		}
	}
	return false
}

func (g *Gossip) SetOutgoingWhitelist(outgoingWhitelist []primitives.MemberId) {
	g.outgoingWhitelist = outgoingWhitelist
}

func (g *Gossip) ClearOutgoingWhitelist() {
	g.SetOutgoingWhitelist(nil)
}

func (g *Gossip) SetIncomingWhitelist(incomingWhitelist []primitives.MemberId) {
	g.incomingWhiteListMemberIds = incomingWhitelist
}

func (g *Gossip) ClearIncomingWhitelist() {
	g.SetIncomingWhitelist(nil)
}

func (g *Gossip) SendToNode(ctx context.Context, targetMemberId primitives.MemberId, consensusRawMessage *leanhelix.ConsensusRawMessage) {
	if g.outgoingWhitelist != nil {
		if !g.inOutgoingWhitelist(targetMemberId) {
			return
		}
	}

	if targetGossip := g.discovery.GetGossipById(targetMemberId); targetGossip != nil {
		targetGossip.OnRemoteMessage(ctx, consensusRawMessage)
	}
	return
}

func (g *Gossip) CountSentMessages(messageType protocol.MessageType) int {
	res := 0
	for _, msg := range g.statsSentMessages {
		if leanhelix.ToConsensusMessage(msg).MessageType() == messageType {
			res++
		}
	}
	return res
}

func (g *Gossip) GetSentMessages(messageType protocol.MessageType) []*leanhelix.ConsensusRawMessage {
	var res []*leanhelix.ConsensusRawMessage
	for _, msg := range g.statsSentMessages {
		if leanhelix.ToConsensusMessage(msg).MessageType() == messageType {
			res = append(res, msg)
		}
	}
	return res
}
