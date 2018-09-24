package gossip

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
)

type Callback func(message lh.MessageTransporter)

type SubscriptionValue struct {
	cb Callback
}

type Gossip struct {
	mock.Mock
	discovery            Discovery
	totalSubscriptions   int
	subscriptions        map[int]*SubscriptionValue
	outgoingWhiteListPKs []lh.PublicKey
	incomingWhiteListPKs []lh.PublicKey
}

func NewGossip(gd Discovery) *Gossip {
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
		if currentPK.Equals(pk) {
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
		if currentPK.Equals(pk) {
			return true
		}
	}
	return false
}

func (g *Gossip) onRemoteMessage(message lh.MessageTransporter) {
	for _, s := range g.subscriptions {
		if !g.inIncomingWhitelist(message.SenderPublicKey()) {
			return
		}
		s.cb(message)
	}
}

func (g *Gossip) subscribe(cb Callback) int {
	g.totalSubscriptions++
	g.subscriptions[g.totalSubscriptions] = &SubscriptionValue{
		cb,
	}
	return g.totalSubscriptions
}

func (g *Gossip) unsubscribe(subscriptionToken int) {
	delete(g.subscriptions, subscriptionToken)
}

func (g *Gossip) unicast(pk lh.PublicKey, message lh.MessageTransporter) {
	if !g.inOutgoingWhitelist(pk) {
		return
	}
	if targetGossip, ok := g.discovery.GetGossipByPK(pk); ok {
		targetGossip.onRemoteMessage(message)
	}
}

func (g *Gossip) Multicast(targetIds []lh.PublicKey, message lh.MessageTransporter) {
	for _, targetId := range targetIds {
		g.unicast(targetId, message)
	}
}
