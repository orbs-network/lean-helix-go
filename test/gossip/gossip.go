package gossip

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"sort"
)

type SubscriptionValue struct {
	cb func(ctx context.Context, message lh.ConsensusRawMessage)
}

type outgoingMessage struct {
	target  MemberId
	message lh.ConsensusRawMessage
}

type Gossip struct {
	discovery            *Discovery
	outgoingChannelsMap  map[string]chan *outgoingMessage
	totalSubscriptions   int
	subscriptions        map[int]*SubscriptionValue
	outgoingWhitelist    []MemberId
	incomingWhiteListPKs []MemberId
	statsSentMessages    []lh.ConsensusRawMessage
}

func NewGossip(discovery *Discovery) *Gossip {
	return &Gossip{
		discovery:            discovery,
		outgoingChannelsMap:  make(map[string]chan *outgoingMessage),
		totalSubscriptions:   0,
		subscriptions:        make(map[int]*SubscriptionValue),
		outgoingWhitelist:    nil,
		incomingWhiteListPKs: nil,
		statsSentMessages:    []lh.ConsensusRawMessage{},
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

func (g *Gossip) RequestOrderedCommittee(ctx context.Context, blockHeight BlockHeight, seed uint64, maxCommitteeSize uint32) []MemberId {
	result := g.discovery.AllGossipsPublicKeys()
	sort.Slice(result, func(i, j int) bool {
		return result[i].KeyForMap() < result[j].KeyForMap()
	})
	return result
}

func (g *Gossip) IsMember(pk MemberId) bool {
	return g.discovery.GetGossipByPK(pk) != nil
}

func (g *Gossip) getOutgoingChannelByTarget(ctx context.Context, target MemberId) chan *outgoingMessage {
	channel := g.outgoingChannelsMap[target.String()]
	if channel == nil {
		channel = make(chan *outgoingMessage, 100)
		g.outgoingChannelsMap[target.String()] = channel
		go g.messageSenderLoop(ctx, channel)
	}

	return channel
}

func (g *Gossip) SendMessage(ctx context.Context, targets []MemberId, message lh.ConsensusRawMessage) {
	g.statsSentMessages = append(g.statsSentMessages, message)
	for _, target := range targets {
		channel := g.getOutgoingChannelByTarget(ctx, target)
		channel <- &outgoingMessage{target, message}
	}
}

func (g *Gossip) RegisterOnMessage(cb func(ctx context.Context, message lh.ConsensusRawMessage)) int {
	g.totalSubscriptions++
	g.subscriptions[g.totalSubscriptions] = &SubscriptionValue{cb}
	return g.totalSubscriptions
}

func (g *Gossip) UnregisterOnMessage(subscriptionToken int) {
	delete(g.subscriptions, subscriptionToken)
}

func (g *Gossip) OnRemoteMessage(ctx context.Context, rawMessage lh.ConsensusRawMessage) {
	for _, s := range g.subscriptions {
		if g.incomingWhiteListPKs != nil {
			senderPublicKey := rawMessage.ToConsensusMessage().SenderPublicKey()
			if !g.inIncomingWhitelist(senderPublicKey) {
				continue
			}
		}
		s.cb(ctx, rawMessage)
	}
}

func (g *Gossip) inIncomingWhitelist(publicKey MemberId) bool {
	for _, currentPK := range g.incomingWhiteListPKs {
		if currentPK.Equal(publicKey) {
			return true
		}
	}
	return false
}

func (g *Gossip) inOutgoingWhitelist(pk MemberId) bool {
	for _, currentPK := range g.outgoingWhitelist {
		if currentPK.Equal(pk) {
			return true
		}
	}
	return false
}

func (g *Gossip) SetOutgoingWhitelist(outgoingWhitelist []MemberId) {
	g.outgoingWhitelist = outgoingWhitelist
}

func (g *Gossip) ClearOutgoingWhitelist() {
	g.SetOutgoingWhitelist(nil)
}

func (g *Gossip) SetIncomingWhitelist(incomingWhitelist []MemberId) {
	g.incomingWhiteListPKs = incomingWhitelist
}

func (g *Gossip) ClearIncomingWhitelist() {
	g.SetIncomingWhitelist(nil)
}

func (g *Gossip) SendToNode(ctx context.Context, targetPublicKey MemberId, consensusRawMessage lh.ConsensusRawMessage) {
	if g.outgoingWhitelist != nil {
		if !g.inOutgoingWhitelist(targetPublicKey) {
			return
		}
	}

	if targetGossip := g.discovery.GetGossipByPK(targetPublicKey); targetGossip != nil {
		targetGossip.OnRemoteMessage(ctx, consensusRawMessage)
	}
	return
}

func (g *Gossip) CountSentMessages(messageType protocol.MessageType) int {
	res := 0
	for _, msg := range g.statsSentMessages {
		if msg.ToConsensusMessage().MessageType() == messageType {
			res++
		}
	}
	return res
}

func (g *Gossip) GetSentMessages(messageType protocol.MessageType) []lh.ConsensusRawMessage {
	var res []lh.ConsensusRawMessage
	for _, msg := range g.statsSentMessages {
		if msg.ToConsensusMessage().MessageType() == messageType {
			res = append(res, msg)
		}
	}
	return res
}
