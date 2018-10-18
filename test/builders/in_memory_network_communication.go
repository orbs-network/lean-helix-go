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

func (comm *InMemoryNetworkCommunication) Send(targets []Ed25519PublicKey, message lh.ConsensusRawMessage) error {
	panic("implement me")
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

func (comm *InMemoryNetworkCommunication) RequestOrderedCommittee(seed uint64) []Ed25519PublicKey {
	return comm.discovery.AllGossipsPKs()
}

func (comm *InMemoryNetworkCommunication) IsMember(pk Ed25519PublicKey) bool {
	_, ok := comm.discovery.GetGossipByPK(pk)
	return ok
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
