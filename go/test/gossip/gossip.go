package gossip

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

type GossipCallback func(Message)

type SubscriptionValue struct {
	cb GossipCallback
}

type Gossip struct {
	discovery            *GossipDiscovery
	totalSubscriptions   int
	subscriptions        map[int]*SubscriptionValue
	outgoingWhiteListPKs []lh.PublicKey
	incomingWhiteListPKs []lh.PublicKey
}

func NewGossip(gd *GossipDiscovery) *Gossip {
	return &Gossip{
		discovery:            gd,
		totalSubscriptions:   0,
		subscriptions:        make(map[int]*SubscriptionValue),
		outgoingWhiteListPKs: nil,
		incomingWhiteListPKs: nil,
	}
}

func (g *Gossip) inIncomingWhitelist(pk lh.PublicKey) bool {
	if g.incomingWhiteListPKs == nil {
		return false
	}
	for _, currentPK := range g.incomingWhiteListPKs {
		if currentPK == pk {
			return true
		}
	}
	return false
}

func (g *Gossip) inOutgoingWhitelist(pk lh.PublicKey) bool {
	if g.outgoingWhiteListPKs == nil {
		return false
	}
	for _, currentPK := range g.outgoingWhiteListPKs {
		if currentPK == pk {
			return true
		}
	}
	return false
}

func (g *Gossip) onRemoteMessage(message lh.MessageContent) {
	for _, s := range g.subscriptions {
		if !g.inIncomingWhitelist(message.SignaturePair.SignerPublicKey) {
			return
		}
		s.cb(message)
	}
}

func (g *Gossip) subscribe(cb GossipCallback) int {
	g.totalSubscriptions++
	g.subscriptions[g.totalSubscriptions] = &SubscriptionValue{
		cb,
	}
	return g.totalSubscriptions
}

func (g *Gossip) unsubscribe(subscriptionToken int) {
	delete(g.subscriptions, subscriptionToken)
}

func (g *Gossip) unicast(pk lh.PublicKey, message Message) {
	if !g.inOutgoingWhitelist(pk) {
		return
	}
	if targetGossip := g.discovery.getGossipByPK(pk); targetGossip != nil {
		targetGossip.onRemoteMessage(message)
	}
}

func (g *Gossip) multicast(targetIds []string, message Message) {
	for _, targetId := range targetIds {
		unicast(targetId, message)
	}
}
