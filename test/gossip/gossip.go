package gossip

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/types"
)

type Callback func(message leanhelix.Message)

type SubscriptionValue struct {
	cb Callback
}

type Gossip struct {
	discovery            *discovery
	totalSubscriptions   int
	subscriptions        map[int]*SubscriptionValue
	outgoingWhiteListPKs []types.PublicKey
	incomingWhiteListPKs []types.PublicKey
}

func NewGossip(gd *discovery) *Gossip {
	return &Gossip{
		discovery:            gd,
		totalSubscriptions:   0,
		subscriptions:        make(map[int]*SubscriptionValue),
		outgoingWhiteListPKs: nil,
		incomingWhiteListPKs: nil,
	}
}

func (g *Gossip) inIncomingWhitelist(pk types.PublicKey) bool {
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

func (g *Gossip) inOutgoingWhitelist(pk types.PublicKey) bool {
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

func (g *Gossip) onRemoteMessage(message leanhelix.Message) {
	for _, s := range g.subscriptions {
		if !g.inIncomingWhitelist(message.SignaturePair().SignerPublicKey) {
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

func (g *Gossip) unicast(pk types.PublicKey, message leanhelix.Message) {
	if !g.inOutgoingWhitelist(pk) {
		return
	}
	if targetGossip, ok := g.discovery.GetGossipByPK(pk); ok {
		targetGossip.onRemoteMessage(message)
	}
}

func (g *Gossip) multicast(targetIds []types.PublicKey, message leanhelix.Message) {
	for _, targetId := range targetIds {
		g.unicast(targetId, message)
	}
}
