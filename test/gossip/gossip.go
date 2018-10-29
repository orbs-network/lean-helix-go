package gossip

import (
	"context"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type SubscriptionValue struct {
	cb func(ctx context.Context, message lh.ConsensusRawMessage)
}

type Gossip struct {
	discovery            Discovery
	publicKey            Ed25519PublicKey
	totalSubscriptions   int
	subscriptions        map[int]*SubscriptionValue
	outgoingWhitelist    []Ed25519PublicKey
	incomingWhiteListPKs []Ed25519PublicKey
	statsSentMessages    []lh.ConsensusRawMessage
}

func NewGossip(gd Discovery, publicKey Ed25519PublicKey) *Gossip {
	return &Gossip{
		discovery:            gd,
		publicKey:            publicKey,
		totalSubscriptions:   0,
		subscriptions:        make(map[int]*SubscriptionValue),
		outgoingWhitelist:    nil,
		incomingWhiteListPKs: nil,
		statsSentMessages:    []lh.ConsensusRawMessage{},
	}
}

func (g *Gossip) RequestOrderedCommittee(seed uint64) []Ed25519PublicKey {
	return g.discovery.AllGossipsPublicKeys()
}

func (g *Gossip) IsMember(pk Ed25519PublicKey) bool {
	return g.discovery.GetGossipByPK(pk) != nil
}

func (g *Gossip) SendMessage(ctx context.Context, targets []Ed25519PublicKey, message lh.ConsensusRawMessage) error {
	g.statsSentMessages = append(g.statsSentMessages, message)
	for _, targetId := range targets {
		if err := g.SendToNode(ctx, targetId, message); err != nil {
			return err
		}
	}
	return nil
}

func (g *Gossip) inIncomingWhitelist(pk Ed25519PublicKey) bool {
	if g.incomingWhiteListPKs == nil {
		return true
	}
	for _, currentPK := range g.incomingWhiteListPKs {
		if currentPK.Equal(pk) {
			return true
		}
	}
	return false
}

func (g *Gossip) inOutgoingWhitelist(pk Ed25519PublicKey) bool {
	if g.outgoingWhitelist == nil {
		return true
	}
	for _, currentPK := range g.outgoingWhitelist {
		if currentPK.Equal(pk) {
			return true
		}
	}
	return false
}

func (g *Gossip) onRemoteMessage(ctx context.Context, targetPublicKey Ed25519PublicKey, rawMessage lh.ConsensusRawMessage) {
	for _, s := range g.subscriptions {

		if !g.inIncomingWhitelist(targetPublicKey) {
			return
		}
		s.cb(ctx, rawMessage)
	}
}

func (g *Gossip) RegisterOnMessage(cb func(ctx context.Context, message lh.ConsensusRawMessage)) int {
	g.totalSubscriptions++
	g.subscriptions[g.totalSubscriptions] = &SubscriptionValue{
		cb,
	}
	return g.totalSubscriptions
}

func (g *Gossip) UnregisterOnMessage(subscriptionToken int) {
	delete(g.subscriptions, subscriptionToken)
}

func (g *Gossip) SendToNode(ctx context.Context, targetPublicKey Ed25519PublicKey, message lh.ConsensusRawMessage) error {
	if !g.inOutgoingWhitelist(targetPublicKey) {
		return fmt.Errorf("PK %s not in outgoing whitelist", targetPublicKey)
	}
	if targetGossip := g.discovery.GetGossipByPK(targetPublicKey); targetGossip != nil {
		targetGossip.onRemoteMessage(ctx, targetPublicKey, message)
	}
	return nil
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

func (g *Gossip) StatsNumSentMessages(predicate func(msg interface{}) bool) int {
	res := 0
	for _, msg := range g.statsSentMessages {
		if predicate(msg) {
			res++
		}
	}
	return res
}
