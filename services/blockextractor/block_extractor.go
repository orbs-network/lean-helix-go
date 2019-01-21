package blockextractor

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sort"
)

func GetLatestBlockFromViewChangeMessages(messages []*interfaces.ViewChangeMessage) (interfaces.Block, primitives.BlockHash) {
	if len(messages) == 0 {
		return nil, nil
	}
	messagesWithBlock := keepOnlyMessagesWithBlock(messages)
	if len(messagesWithBlock) == 0 {
		return nil, nil
	}
	sortedMessagesWithBlock := sortMessagesByDescendingViewOfPreparedProofPPM(messagesWithBlock)
	latestVC := sortedMessagesWithBlock[0]
	return latestVC.Block(), latestVC.Content().SignedHeader().PreparedProof().PrepareBlockRef().BlockHash()
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
