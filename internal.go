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
	var message ConsensusMessage
	lhContentReader := protocol.LeanhelixContentReader(c.Content())

	if lhContentReader.IsMessagePreprepareMessage() {
		message = &PreprepareMessage{
			content: lhContentReader.PreprepareMessage(),
			block:   c.Block(),
		}
	}

	if lhContentReader.IsMessagePrepareMessage() {
		message = &PrepareMessage{
			content: lhContentReader.PrepareMessage(),
		}
	}

	if lhContentReader.IsMessageCommitMessage() {
		message = &CommitMessage{
			content: lhContentReader.CommitMessage(),
		}
		return message
	}

	if lhContentReader.IsMessageViewChangeMessage() {
		message = &ViewChangeMessage{
			content: lhContentReader.ViewChangeMessage(),
			block:   c.Block(),
		}
	}

	if lhContentReader.IsMessageNewViewMessage() {
		message = &NewViewMessage{
			content: lhContentReader.NewViewMessage(),
			block:   c.Block(),
		}
	}
	return message // handle with error
}
