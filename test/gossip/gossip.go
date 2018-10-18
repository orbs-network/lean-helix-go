package gossip

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type Callback func(message lh.ConsensusMessage)

type SubscriptionValue struct {
	cb Callback
}

type Gossip struct {
	mock.Mock
	discovery            Discovery
	publicKey            Ed25519PublicKey
	totalSubscriptions   int
	subscriptions        map[int]*SubscriptionValue
	outgoingWhiteListPKs []Ed25519PublicKey
	incomingWhiteListPKs []Ed25519PublicKey
}

func NewGossip(gd Discovery, publicKey Ed25519PublicKey) *Gossip {
	return &Gossip{
		discovery:            gd,
		publicKey:            publicKey,
		totalSubscriptions:   0,
		subscriptions:        make(map[int]*SubscriptionValue),
		outgoingWhiteListPKs: nil,
		incomingWhiteListPKs: nil,
	}
}

func (g *Gossip) inIncomingWhitelist(pk Ed25519PublicKey) bool {
	if g.incomingWhiteListPKs == nil {
		return false
	}
	for _, currentPK := range g.incomingWhiteListPKs {
		if currentPK.Equal(pk) {
			return true
		}
	}
	return false
}

func (g *Gossip) inOutgoingWhitelist(pk Ed25519PublicKey) bool {
	if g.outgoingWhiteListPKs == nil {
		return false
	}
	for _, currentPK := range g.outgoingWhiteListPKs {
		if currentPK.Equal(pk) {
			return true
		}
	}
	return false
}

func (g *Gossip) onRemoteMessage(message lh.ConsensusMessage) {
	for _, s := range g.subscriptions {
		if !g.inIncomingWhitelist(message.SenderPublicKey()) {
			return
		}
		s.cb(message)
	}
}

func (g *Gossip) Subscribe(cb Callback) int {
	g.totalSubscriptions++
	g.subscriptions[g.totalSubscriptions] = &SubscriptionValue{
		cb,
	}
	return g.totalSubscriptions
}

func (g *Gossip) Unsubscribe(subscriptionToken int) {
	delete(g.subscriptions, subscriptionToken)
}

func (g *Gossip) unicast(pk Ed25519PublicKey, message lh.ConsensusMessage) {
	if !g.inOutgoingWhitelist(pk) {
		return
	}
	if targetGossip, ok := g.discovery.GetGossipByPK(pk); ok {
		targetGossip.onRemoteMessage(message)
	}
}

func (g *Gossip) Multicast(targetIds []Ed25519PublicKey, message lh.ConsensusMessage) {
	g.Mock.Called(targetIds, message)
	for _, targetId := range targetIds {
		g.unicast(targetId, message)
	}
}
