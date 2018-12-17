package builders

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type NodeBuilder struct {
	gossip        *gossip.Gossip
	blocksPool    *BlocksPool
	logsToConsole bool
	memberId      primitives.MemberId
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

func (builder *NodeBuilder) ThatIsPartOf(gossip *gossip.Gossip) *NodeBuilder {
	if builder.gossip == nil {
		builder.gossip = gossip
	}
	return builder
}

func (builder *NodeBuilder) WithMemberId(memberId primitives.MemberId) *NodeBuilder {
	if builder.memberId == nil {
		builder.memberId = memberId
	}
	return builder
}

func (builder *NodeBuilder) WithBlocksPool(blocksPool *BlocksPool) *NodeBuilder {
	if builder.blocksPool == nil {
		builder.blocksPool = blocksPool
	}
	return builder
}

func (builder *NodeBuilder) ThatLogsToConsole() *NodeBuilder {
	builder.logsToConsole = true
	return builder
}

func (builder *NodeBuilder) Build() *Node {
	memberId := builder.memberId
	if memberId == nil {
		memberId = primitives.MemberId(fmt.Sprintf("Dummy MemberId"))
	}

	blockUtils := NewMockBlockUtils(builder.blocksPool)
	electionTrigger := NewMockElectionTrigger()
	var logger leanhelix.Logger
	if builder.logsToConsole {
		logger = leanhelix.NewConsoleLogger(memberId.KeyForMap())
	}
	return NewNode(memberId, builder.gossip, blockUtils, electionTrigger, logger)
}
