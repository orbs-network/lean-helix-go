package builders

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetworkBuilder struct {
	NodeCount          int
	logToConsole       bool
	customNodeBuilders []*NodeBuilder
	upcomingBlocks     []leanhelix.Block
	keyManager         leanhelix.KeyManager
	blockUtils         leanhelix.BlockUtils
	communication      leanhelix.Communication
}

func (tb *TestNetworkBuilder) WithNodeCount(nodeCount int) *TestNetworkBuilder {
	tb.NodeCount = nodeCount
	return tb
}

func (tb *TestNetworkBuilder) WithCustomNodeBuilder(nodeBuilder *NodeBuilder) *TestNetworkBuilder {
	tb.customNodeBuilders = append(tb.customNodeBuilders, nodeBuilder)
	return tb
}

func (tb *TestNetworkBuilder) WithBlocks(upcomingBlocks []leanhelix.Block) *TestNetworkBuilder {
	if tb.upcomingBlocks == nil {
		tb.upcomingBlocks = upcomingBlocks
	}
	return tb
}

func (tb *TestNetworkBuilder) LogToConsole() *TestNetworkBuilder {
	tb.logToConsole = true
	return tb
}

func (tb *TestNetworkBuilder) Build() *TestNetwork {
	blocksPool := tb.buildBlocksPool()
	discovery := gossip.NewDiscovery()
	nodes := tb.createNodes(discovery, blocksPool, tb.logToConsole)
	testNetwork := NewTestNetwork(discovery)
	testNetwork.RegisterNodes(nodes)
	return testNetwork
}

func (tb *TestNetworkBuilder) buildBlocksPool() *BlocksPool {
	if tb.upcomingBlocks == nil {
		b1 := CreateBlock(GenesisBlock)
		b2 := CreateBlock(b1)
		b3 := CreateBlock(b2)
		b4 := CreateBlock(b3)

		return NewBlocksPool([]leanhelix.Block{b1, b2, b3, b4})
	} else {
		return NewBlocksPool(tb.upcomingBlocks)
	}
}

func (tb *TestNetworkBuilder) buildNode(
	nodeBuilder *NodeBuilder,
	memberId primitives.MemberId,
	discovery *gossip.Discovery,
	blocksPool *BlocksPool,
	logToConsole bool) *Node {

	gossipInstance := gossip.NewGossip(discovery)
	discovery.RegisterGossip(memberId, gossipInstance)
	membership := gossip.NewMockMembership(memberId, discovery)

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

func (tb *TestNetworkBuilder) createNodes(discovery *gossip.Discovery, blocksPool *BlocksPool, logToConsole bool) []*Node {
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

func (tb *TestNetworkBuilder) WithCommunication(communication leanhelix.Communication) *TestNetworkBuilder {
	tb.communication = communication
	return tb
}

func (tb *TestNetworkBuilder) WithBlockUtils(utils leanhelix.BlockUtils) *TestNetworkBuilder {
	tb.blockUtils = utils
	return tb
}

func (tb *TestNetworkBuilder) WithKeyManager(mgr leanhelix.KeyManager) *TestNetworkBuilder {
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

func ATestNetwork(countOfNodes int, blocksPool ...leanhelix.Block) *TestNetwork {
	return NewTestNetworkBuilder().WithNodeCount(countOfNodes).WithBlocks(blocksPool).Build()
}

func CreateTestNetworkForConsumerTests(
	countOfNodes int,
	spi *leanhelix.LeanHelixSPI,
	blocks []leanhelix.Block,
) *TestNetwork {
	testNetwork := NewTestNetworkBuilder()
	return testNetwork.
		WithNodeCount(countOfNodes).
		WithBlocks(blocks).
		WithCommunication(spi.Comm).
		WithKeyManager(spi.Mgr).
		WithBlockUtils(spi.Utils).
		Build()
}
