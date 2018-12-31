package network

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

type NodeBuilder struct {
	communication *mocks.CommunicationMock
	membership    interfaces.Membership
	blocksPool    *mocks.BlocksPool
	logsToConsole bool
	memberId      primitives.MemberId
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

func (builder *NodeBuilder) CommunicatesVia(communication *mocks.CommunicationMock) *NodeBuilder {
	if builder.communication == nil {
		builder.communication = communication
	}
	return builder
}

func (builder *NodeBuilder) ThatIsPartOf(membership interfaces.Membership) *NodeBuilder {
	if builder.membership == nil {
		builder.membership = membership
	}
	return builder
}

func (builder *NodeBuilder) WithMemberId(memberId primitives.MemberId) *NodeBuilder {
	if builder.memberId == nil {
		builder.memberId = memberId
	}
	return builder
}

func (builder *NodeBuilder) WithBlocksPool(blocksPool *mocks.BlocksPool) *NodeBuilder {
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

	blockUtils := mocks.NewMockBlockUtils(builder.blocksPool)
	electionTrigger := mocks.NewMockElectionTrigger()
	var l interfaces.Logger
	if builder.logsToConsole {
		l = logger.NewConsoleLogger(memberId.KeyForMap())
	}
	return NewNode(builder.membership, builder.communication, blockUtils, electionTrigger, l)
}

func ADummyNode() *Node {
	memberId := primitives.MemberId("Dummy")
	return NewNodeBuilder().
		WithMemberId(memberId).
		ThatIsPartOf(mocks.NewMockMembership(memberId, nil, false)).
		CommunicatesVia(mocks.NewCommunication(nil)).
		Build()
}
