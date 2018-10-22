package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

// Sorting View arrays
type ViewCounters []primitives.View

func (arr ViewCounters) Len() int           { return len(arr) }
func (arr ViewCounters) Swap(i, j int)      { arr[i], arr[j] = arr[j], arr[i] }
func (arr ViewCounters) Less(i, j int) bool { return arr[i] < arr[j] }

type consensusRawMessage struct {
	messageType MessageType
	content     []byte
	block       Block
}

func (c *consensusRawMessage) MessageType() MessageType {
	return c.messageType
}

func (c *consensusRawMessage) Content() []byte {
	return c.content
}

func (c *consensusRawMessage) Block() Block {
	return c.block
}

func CreateConsensusRawMessage(messageType MessageType, content []byte, block Block) ConsensusRawMessage {
	return &consensusRawMessage{
		messageType: messageType,
		content:     content,
		block:       block,
	}
}
