package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

func CreatePreparedMessages(
	leader *Node,
	members []*Node,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PreparedMessages {

	PPMessage := APreprepareMessage(leader.KeyManager, leader.MemberId, blockHeight, view, block)
	PMessages := make([]*leanhelix.PrepareMessage, len(members))

	for i, member := range members {
		PMessages[i] = APrepareMessage(member.KeyManager, member.MemberId, blockHeight, view, block)
	}

	return &leanhelix.PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}
}
