package blockextractor

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"sort"
)

func GetLatestBlockFromViewChangeMessages(messages []*interfaces.ViewChangeMessage) interfaces.Block {
	if len(messages) == 0 {
		return nil
	}
	messagesWithBlock := keepOnlyMessagesWithBlock(messages)
	if len(messagesWithBlock) == 0 {
		return nil
	}
	sortedMessagesWithBlock := sortMessagesByDescendingViewOfPreparedProofPPM(messagesWithBlock)
	return sortedMessagesWithBlock[0].Block()
}

func keepOnlyMessagesWithBlock(msgs []*interfaces.ViewChangeMessage) []*interfaces.ViewChangeMessage {
	messagesWithBlock := make([]*interfaces.ViewChangeMessage, 0, len(msgs))
	for _, msg := range msgs {
		if msg.Block() != nil {
			messagesWithBlock = append(messagesWithBlock, msg)
		}
	}
	return messagesWithBlock
}

func sortMessagesByDescendingViewOfPreparedProofPPM(msgs []*interfaces.ViewChangeMessage) []*interfaces.ViewChangeMessage {
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Content().SignedHeader().PreparedProof().PreprepareBlockRef().View() > msgs[j].Content().SignedHeader().PreparedProof().PreprepareBlockRef().View()
	})
	return msgs
}
