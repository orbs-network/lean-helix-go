package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

// Sorting View arrays
type ViewCounters []primitives.View

func (arr ViewCounters) Len() int           { return len(arr) }
func (arr ViewCounters) Swap(i, j int)      { arr[i], arr[j] = arr[j], arr[i] }
func (arr ViewCounters) Less(i, j int) bool { return arr[i] < arr[j] }

type consensusRawMessage struct {
	content []byte
	block   Block
}

func (c *consensusRawMessage) Content() []byte {
	return c.content
}

func (c *consensusRawMessage) Block() Block {
	return c.block
}

func (c *consensusRawMessage) ToConsensusMessage() ConsensusMessage {
	content := protocol.LeanhelixContentReader(c.Content())

	if content.IsMessagePreprepareMessage() {
		content := protocol.PreprepareContentReader(c.Content())
		return &PreprepareMessage{
			content: content,
			block:   c.Block(),
		}
	}

	if content.IsMessagePrepareMessage() {
		content := protocol.PrepareContentReader(c.Content())
		return &PrepareMessage{
			content: content,
		}
	}

	if content.IsMessagePrepareMessage() {
		content := protocol.CommitContentReader(c.Content())
		return &CommitMessage{
			content: content,
		}
	}

	if content.IsMessagePrepareMessage() {
		content := protocol.ViewChangeMessageContentReader(c.Content())
		return &ViewChangeMessage{
			content: content,
			block:   c.Block(),
		}
	}

	if content.IsMessagePrepareMessage() {
		content := protocol.NewViewMessageContentReader(c.Content())
		return &NewViewMessage{
			content: content,
			block:   c.Block(),
		}
	}

	return nil // handle with error
}

func CreateConsensusRawMessage(content []byte, block Block) ConsensusRawMessage {
	return &consensusRawMessage{
		content: content,
		block:   block,
	}
}
