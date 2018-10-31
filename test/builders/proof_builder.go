package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

func CreatePreparedMessages(
	leader *Node,
	members []*Node,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PreparedMessages {

	mf := leanhelix.NewMessageFactory(leader.KeyManager)
	PPMessage := mf.CreatePreprepareMessage(blockHeight, view, block)

	PMessages := make([]*leanhelix.PrepareMessage, len(members))
	for i, member := range members {
		mf := leanhelix.NewMessageFactory(member.KeyManager)
		PMessages[i] = mf.CreatePrepareMessage(blockHeight, view, block.BlockHash())
	}

	return &leanhelix.PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}
}
