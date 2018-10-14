package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
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

func (comm *InMemoryNetworkCommunication) Send(publicKeys []Ed25519PublicKey, message []byte) error {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendWithBlock(publicKeys []Ed25519PublicKey, message []byte, block lh.Block) error {
	panic("implement me")
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
	case lh.LEAN_HELIX_PREPREPARE:
		if comm.PPCallback != nil {
			comm.PPCallback(message.(lh.PreprepareMessage))
		}
	case lh.LEAN_HELIX_PREPARE:
		if comm.PCallback != nil {
			comm.PCallback(message.(lh.PrepareMessage))
		}
	case lh.LEAN_HELIX_COMMIT:
		if comm.CCallback != nil {
			comm.CCallback(message.(lh.CommitMessage))
		}
	case lh.LEAN_HELIX_VIEW_CHANGE:
		if comm.VCCallback != nil {
			comm.VCCallback(message.(lh.ViewChangeMessage))
		}
	case lh.LEAN_HELIX_NEW_VIEW:
		if comm.NVCallback != nil {
			comm.NVCallback(message.(lh.NewViewMessage))
		}
	}
}

func (comm *InMemoryNetworkCommunication) SendToMembers(publicKeys []Ed25519PublicKey, messageType string, message []lh.MessageTransporter) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) RequestOrderedCommittee(seed uint64) []Ed25519PublicKey {
	return comm.discovery.AllGossipsPKs()
}

func (comm *InMemoryNetworkCommunication) IsMember(pk Ed25519PublicKey) bool {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendPreprepare(pks []Ed25519PublicKey, message lh.PreprepareMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendPrepare(pks []Ed25519PublicKey, message lh.PrepareMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendCommit(pks []Ed25519PublicKey, message lh.CommitMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendViewChange(pk Ed25519PublicKey, message lh.ViewChangeMessage) {
	panic("implement me")
}

func (comm *InMemoryNetworkCommunication) SendNewView(pks []Ed25519PublicKey, message lh.NewViewMessage) {
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
