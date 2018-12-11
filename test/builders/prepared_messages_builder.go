package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

func CreatePreparedMessages(
	leaderKeyManager leanhelix.KeyManager,
	membersKeyManagers []leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PreparedMessages {

	PPMessage := APreprepareMessage(leaderKeyManager, blockHeight, view, block)
	PMessages := make([]*leanhelix.PrepareMessage, len(membersKeyManagers))

	for i, member := range membersKeyManagers {
		PMessages[i] = APrepareMessage(member, blockHeight, view, block)
	}

	return &leanhelix.PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}
}
