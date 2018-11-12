package gossip

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"sort"
)

type SubscriptionValue struct {
	cb func(ctx context.Context, message lh.ConsensusRawMessage)
}

type Gossip struct {
	discovery            *Discovery
	totalSubscriptions   int
	subscriptions        map[int]*SubscriptionValue
	outgoingWhitelist    []Ed25519PublicKey
	incomingWhiteListPKs []Ed25519PublicKey
	statsSentMessages    []lh.ConsensusRawMessage
}

func NewGossip(discovery *Discovery) *Gossip {
	return &Gossip{
		discovery:            discovery,
		totalSubscriptions:   0,
		subscriptions:        make(map[int]*SubscriptionValue),
		outgoingWhitelist:    nil,
		incomingWhiteListPKs: nil,
		statsSentMessages:    []lh.ConsensusRawMessage{},
	}
}

func (g *Gossip) RequestOrderedCommittee(seed uint64) []Ed25519PublicKey {
	result := g.discovery.AllGossipsPublicKeys()
	sort.Slice(result, func(i, j int) bool {
		return result[i].KeyForMap() < result[j].KeyForMap()
	})
	return result
}

func (g *Gossip) IsMember(pk Ed25519PublicKey) bool {
	return g.discovery.GetGossipByPK(pk) != nil
}

func (g *Gossip) SendMessage(ctx context.Context, targets []Ed25519PublicKey, message lh.ConsensusRawMessage) {
	g.statsSentMessages = append(g.statsSentMessages, message)
	for _, targetId := range targets {
		g.SendToNode(ctx, targetId, message)
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

func (g *Gossip) onRemoteMessage(ctx context.Context, rawMessage lh.ConsensusRawMessage) {
	for _, s := range g.subscriptions {
		if g.incomingWhiteListPKs != nil {
			senderPublicKey := rawMessage.ToConsensusMessage().SenderPublicKey()
			if !g.inIncomingWhitelist(senderPublicKey) {
				continue
			}
		}
		go s.cb(ctx, rawMessage)
	}
}

func (g *Gossip) inIncomingWhitelist(publicKey Ed25519PublicKey) bool {
	for _, currentPK := range g.incomingWhiteListPKs {
		if currentPK.Equal(publicKey) {
			return true
		}
	}
	return false
}

func (g *Gossip) inOutgoingWhitelist(pk Ed25519PublicKey) bool {
	for _, currentPK := range g.outgoingWhitelist {
		if currentPK.Equal(pk) {
			return true
		}
	}
	return false
}

func (g *Gossip) SetOutgoingWhitelist(outgoingWhitelist []Ed25519PublicKey) {
	g.outgoingWhitelist = outgoingWhitelist
}

func (g *Gossip) ClearOutgoingWhitelist(outgoingWhitelist []Ed25519PublicKey) {
	g.SetOutgoingWhitelist(nil)
}

func (g *Gossip) SetIncomingWhitelist(incomingWhitelist []Ed25519PublicKey) {
	g.incomingWhiteListPKs = incomingWhitelist
}

func (g *Gossip) ClearIncomingWhitelist(incomingWhitelist []Ed25519PublicKey) {
	g.SetIncomingWhitelist(nil)
}

func (g *Gossip) SendToNode(ctx context.Context, targetPublicKey Ed25519PublicKey, consensusRawMessage lh.ConsensusRawMessage) {
	if g.outgoingWhitelist != nil {
		if !g.inOutgoingWhitelist(targetPublicKey) {
			return
		}
	}

	if targetGossip := g.discovery.GetGossipByPK(targetPublicKey); targetGossip != nil {
		targetGossip.onRemoteMessage(ctx, consensusRawMessage)
	}
	return
}

func (g *Gossip) StatsNumSentMessages(predicate func(msg interface{}) bool) int {
	res := 0
	for _, msg := range g.statsSentMessages {
		if predicate(msg) {
			res++
		}
	}
	return res
}
