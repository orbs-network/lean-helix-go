package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type InMemoryNetworkCommunication struct {
	PPCallback func(message lh.PreprepareMessage)
	PCallback  func(message lh.PrepareMessage)
	CCallback  func(message lh.CommitMessage)
	VCCallback func(message lh.ViewChangeMessage)
	NVCallback func(message lh.NewViewMessage)

	discovery gossip.Discovery
	gossip    *gossip.Gossip
}

func NewInMemoryNetworkCommunication(discovery gossip.Discovery, gossip *gossip.Gossip) *InMemoryNetworkCommunication {

	comm := &InMemoryNetworkCommunication{
		discovery: discovery,
		gossip:    gossip,
	}

	subscribeFunc := func(message lh.MessageTransporter) {
		comm.onGossipMessage(message)
	}

	comm.gossip.Subscribe(subscribeFunc)

	return comm
}

func (comm *InMemoryNetworkCommunication) onGossipMessage(message lh.MessageTransporter) {
	switch message.MessageType() {
	case lh.MESSAGE_TYPE_PREPREPARE:
		if comm.PPCallback != nil {
			comm.PPCallback(message.(lh.PreprepareMessage))
		}
	case lh.MESSAGE_TYPE_PREPARE:
		if comm.PCallback != nil {
			comm.PCallback(message.(lh.PrepareMessage))
		}
	case lh.MESSAGE_TYPE_COMMIT:
		if comm.CCallback != nil {
			comm.CCallback(message.(lh.CommitMessage))
		}
	case lh.MESSAGE_TYPE_VIEW_CHANGE:
		if comm.VCCallback != nil {
			comm.VCCallback(message.(lh.ViewChangeMessage))
		}
	case lh.MESSAGE_TYPE_NEW_VIEW:
		if comm.NVCallback != nil {
			comm.NVCallback(message.(lh.NewViewMessage))
		}
	}
}

func (comm *InMemoryNetworkCommunication) SendToMembers(publicKeys []lh.PublicKey, messageType string, message []lh.MessageTransporter) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) GetMembersPKs(seed uint64) []lh.PublicKey {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) IsMember(pk lh.PublicKey) bool {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendPreprepare(pks []lh.PublicKey, message lh.PreprepareMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendPrepare(pks []lh.PublicKey, message lh.PrepareMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendCommit(pks []lh.PublicKey, message lh.CommitMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendViewChange(pk lh.PublicKey, message lh.ViewChangeMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendNewView(pks []lh.PublicKey, message lh.NewViewMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) RegisterToPreprepare(cb func(message lh.PreprepareMessage)) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) RegisterToPrepare(cb func(message lh.PrepareMessage)) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) RegisterToCommit(cb func(message lh.CommitMessage)) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) RegisterToViewChange(cb func(message lh.ViewChangeMessage)) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) RegisterToNewView(cb func(message lh.NewViewMessage)) {
	panic("implement me")
}
