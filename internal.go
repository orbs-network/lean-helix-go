package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

// Sorting View arrays
type ViewCounters []primitives.View

func (arr ViewCounters) Len() int           { return len(arr) }
func (arr ViewCounters) Swap(i, j int)      { arr[i], arr[j] = arr[j], arr[i] }
func (arr ViewCounters) Less(i, j int) bool { return arr[i] < arr[j] }

type consensusMessage struct {
	content []byte
	block   Block
}

func (c *consensusMessage) Content() []byte {
	return c.content
}

func (c *consensusMessage) Block() Block {
	return c.block
}

func CreateConsensusMessage(content []byte, block Block) ConsensusMessage {
	return &consensusMessage{
		content: content,
		block:   block,
	}
}
