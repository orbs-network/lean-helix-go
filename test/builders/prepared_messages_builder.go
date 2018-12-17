package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

func CreatePreparedMessages(
	leaderKeyManager leanhelix.KeyManager,
	leaderMemberId primitives.MemberId,
	membersKeyManagers []leanhelix.KeyManager,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block leanhelix.Block) *leanhelix.PreparedMessages {

	PPMessage := APreprepareMessage(leaderKeyManager, leaderMemberId, blockHeight, view, block)
	PMessages := make([]*leanhelix.PrepareMessage, len(membersKeyManagers))

	for i, member := range membersKeyManagers {
		PMessages[i] = APrepareMessage(member, leaderMemberId, blockHeight, view, block)
	}

	return &leanhelix.PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}
}
