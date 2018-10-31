package leanhelix

import "sort"

func GetLatestBlockFromViewChangeMessages(messages []*ViewChangeMessage) Block {
	if len(messages) == 0 {
		return nil
	}
	messagesWithBlock := keepOnlyMessagesWithBlock(messages)
	if len(messagesWithBlock) == 0 {
		return nil
	}
	sortedMessagesWithBlock := sortMessagesByDescendingViewOfPreparedProofPPM(messagesWithBlock)
	return sortedMessagesWithBlock[0].block
}

func keepOnlyMessagesWithBlock(messages []*ViewChangeMessage) []*ViewChangeMessage {
	messagesWithBlock := make([]*ViewChangeMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.block != nil {
			messagesWithBlock = append(messagesWithBlock, msg)
		}
	}
	return messagesWithBlock
}

func sortMessagesByDescendingViewOfPreparedProofPPM(messages []*ViewChangeMessage) []*ViewChangeMessage {
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].content.SignedHeader().PreparedProof().PreprepareBlockRef().View() > messages[j].content.SignedHeader().PreparedProof().PreprepareBlockRef().View()
	})
	return messages
}
