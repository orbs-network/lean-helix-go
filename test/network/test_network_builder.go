package network

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

type TestNetworkBuilder struct {
	NodeCount              int
	logToConsole           bool
	customNodeBuilders     []*NodeBuilder
	upcomingBlocks         []interfaces.Block
	keyManager             interfaces.KeyManager
	blockUtils             interfaces.BlockUtils
	communication          interfaces.Communication
	orderCommitteeByHeight bool
}

func (tb *TestNetworkBuilder) WithNodeCount(nodeCount int) *TestNetworkBuilder {
	tb.NodeCount = nodeCount
	return tb
}

func (tb *TestNetworkBuilder) WithCustomNodeBuilder(nodeBuilder *NodeBuilder) *TestNetworkBuilder {
	tb.customNodeBuilders = append(tb.customNodeBuilders, nodeBuilder)
	return tb
}

func (tb *TestNetworkBuilder) WithBlocks(upcomingBlocks []interfaces.Block) *TestNetworkBuilder {
	if tb.upcomingBlocks == nil {
		tb.upcomingBlocks = upcomingBlocks
	}
	return tb
}

func (tb *TestNetworkBuilder) LogToConsole() *TestNetworkBuilder {
	tb.logToConsole = true
	return tb
}

func (tb *TestNetworkBuilder) OrderCommitteeByHeight() *TestNetworkBuilder {
	tb.orderCommitteeByHeight = true
	return tb
}

func (tb *TestNetworkBuilder) Build() *TestNetwork {
	blocksPool := tb.buildBlocksPool()
	discovery := mocks.NewDiscovery()
	nodes := tb.createNodes(discovery, blocksPool, tb.logToConsole)
	testNetwork := NewTestNetwork(discovery)
	testNetwork.RegisterNodes(nodes)
	return testNetwork
}

func (tb *TestNetworkBuilder) buildBlocksPool() *mocks.BlocksPool {
	if tb.upcomingBlocks == nil {
		b1 := mocks.ABlock(interfaces.GenesisBlock)
		b2 := mocks.ABlock(b1)
		b3 := mocks.ABlock(b2)
		b4 := mocks.ABlock(b3)

		return mocks.NewBlocksPool([]interfaces.Block{b1, b2, b3, b4})
	} else {
		return mocks.NewBlocksPool(tb.upcomingBlocks)
	}
}

func (tb *TestNetworkBuilder) buildNode(
	nodeBuilder *NodeBuilder,
	memberId primitives.MemberId,
	discovery *mocks.Discovery,
	blocksPool *mocks.BlocksPool,
	logToConsole bool) *Node {

	gossipInstance := mocks.NewGossip(discovery)
	discovery.RegisterGossip(memberId, gossipInstance)
	membership := mocks.NewMockMembership(memberId, discovery, tb.orderCommitteeByHeight)

	b := nodeBuilder.
		CommunicatesVia(gossipInstance).
		ThatIsPartOf(membership).
		WithBlocksPool(blocksPool).
		WithMemberId(memberId)

	if logToConsole {
		b.ThatLogsToConsole()
	}
	return b.Build()
}

func (tb *TestNetworkBuilder) createNodes(discovery *mocks.Discovery, blocksPool *mocks.BlocksPool, logToConsole bool) []*Node {
	var nodes []*Node
	for i := 0; i < tb.NodeCount; i++ {
		nodeBuilder := NewNodeBuilder()
		memberId := primitives.MemberId(fmt.Sprintf("Node %d", i))
		node := tb.buildNode(nodeBuilder, memberId, discovery, blocksPool, logToConsole)
		nodes = append(nodes, node)
	}

	for i, customBuilder := range tb.customNodeBuilders {
		memberId := primitives.MemberId(fmt.Sprintf("Custom-Node %d", i))
		node := tb.buildNode(customBuilder, memberId, discovery, blocksPool, logToConsole)
		nodes = append(nodes, node)
	}

	return nodes
}

func (tb *TestNetworkBuilder) WithCommunication(communication interfaces.Communication) *TestNetworkBuilder {
	tb.communication = communication
	return tb
}

func (tb *TestNetworkBuilder) WithBlockUtils(utils interfaces.BlockUtils) *TestNetworkBuilder {
	tb.blockUtils = utils
	return tb
}

func (tb *TestNetworkBuilder) WithKeyManager(mgr interfaces.KeyManager) *TestNetworkBuilder {
	tb.keyManager = mgr
	return tb
}

func NewTestNetworkBuilder() *TestNetworkBuilder {
	return &TestNetworkBuilder{
		NodeCount:          0,
		customNodeBuilders: nil,
		upcomingBlocks:     nil,
	}
}

func ABasicTestNetwork() *TestNetwork {
	return ATestNetwork(4)
}

func ATestNetwork(countOfNodes int, blocksPool ...interfaces.Block) *TestNetwork {
	return NewTestNetworkBuilder().WithNodeCount(countOfNodes).WithBlocks(blocksPool).Build()
}
