package builders

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetworkBuilder struct {
	NodeCount            int
	logToConsole         bool
	customNodeBuilders   []*NodeBuilder
	upcomingBlocks       []leanhelix.Block
	keyManager           leanhelix.KeyManager
	blockUtils           leanhelix.BlockUtils
	networkCommunication leanhelix.NetworkCommunication
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
	discovery := gossip.NewGossipDiscovery()
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
	publicKey primitives.MemberId,
	discovery *gossip.Discovery,
	blocksPool *BlocksPool,
	logToConsole bool) *Node {

	gossip := gossip.NewGossip(discovery)
	discovery.RegisterGossip(publicKey, gossip)

	b := nodeBuilder.
		ThatIsPartOf(gossip).
		WithBlocksPool(blocksPool).
		WithPublicKey(publicKey)

	if logToConsole {
		b.ThatLogsToConsole()
	}
	return b.Build()
}

func (tb *TestNetworkBuilder) createNodes(discovery *gossip.Discovery, blocksPool *BlocksPool, logToConsole bool) []*Node {
	var nodes []*Node
	for i := 0; i < tb.NodeCount; i++ {
		nodeBuilder := NewNodeBuilder()
		publicKey := primitives.MemberId(fmt.Sprintf("Node %d", i))
		node := tb.buildNode(nodeBuilder, publicKey, discovery, blocksPool, logToConsole)
		nodes = append(nodes, node)
	}

	for i, customBuilder := range tb.customNodeBuilders {
		publicKey := primitives.MemberId(fmt.Sprintf("Custom-Node %d", i))
		node := tb.buildNode(customBuilder, publicKey, discovery, blocksPool, logToConsole)
		nodes = append(nodes, node)
	}

	return nodes
}

func (tb *TestNetworkBuilder) WithNetworkCommunication(comm leanhelix.NetworkCommunication) *TestNetworkBuilder {
	tb.networkCommunication = comm
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
		WithNetworkCommunication(spi.Comm).
		WithKeyManager(spi.Mgr).
		WithBlockUtils(spi.Utils).
		Build()
}
