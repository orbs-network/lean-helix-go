package builders

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Sender interface {
	GetKeyManager() interfaces.KeyManager
	GetMemberId() primitives.MemberId
}

func CreatePreparedMessages(
	networkId primitives.NetworkId,
	leader Sender,
	members []Sender,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block) *preparedmessages.PreparedMessages {

	PPMessage := APreprepareMessage(networkId, leader.GetKeyManager(), leader.GetMemberId(), blockHeight, view, block)
	PMessages := make([]*interfaces.PrepareMessage, len(members))

	for i, member := range members {
		PMessages[i] = APrepareMessage(networkId, member.GetKeyManager(), member.GetMemberId(), blockHeight, view, block)
	}

	return &preparedmessages.PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}
}
