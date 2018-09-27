package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/mock"
)

type TestNetwork struct {
	mock.Mock
	Nodes      []*Node
	BlockUtils *MockBlockUtils
	Transport  *MockNetworkCommunication
	discovery  gossip.Discovery
}

type TestNetworkBuilder struct {
	logger          lh.Logger
	customNodes     []NodeBuilder
	electionTrigger lh.ElectionTrigger
	blockUtils      lh.BlockUtils
	blocksPool      []lh.Block
	nodeCount       int
}

func (builder *TestNetworkBuilder) GettingBlocksVia(utils lh.BlockUtils) *TestNetworkBuilder {
	builder.blockUtils = utils
	return builder
}

func (builder *TestNetworkBuilder) WithNodeCount(nodeCount int) *TestNetworkBuilder {
	builder.nodeCount = nodeCount
	return builder
}

func (builder *TestNetworkBuilder) ThatLogsToCustomLogger(logger lh.Logger) *TestNetworkBuilder {
	builder.logger = logger
	return builder
}

func (builder *TestNetworkBuilder) Build() *TestNetwork {

	discovery := gossip.NewGossipDiscovery()
	nodes := createNodes(builder.nodeCount, discovery)

	testNet := &TestNetwork{
		Nodes: nodes,
		//BlockUtils: builder.blockUtils,
		//Transport:  builder.transport,
		discovery: discovery,
	}

	// TODO Why we need this?? it does nothing on TS code
	//testNet.registerNodes()

	return testNet
}

func CreateSimpleTestNetwork(nodeCount int, blocksPool []lh.Block) *TestNetwork {

	b1 := CreateBlock(GenesisBlock)
	b2 := CreateBlock(b1)
	b3 := CreateBlock(b2)
	b4 := CreateBlock(b3)

	var blocks []lh.Block
	if blocksPool != nil {
		blocks = blocksPool
	} else {
		blocks = []lh.Block{b1, b2, b3, b4}
	}

	mockBlockUtils := NewMockBlockUtils(blocks)

	testNet := CreateTestNetworkBuilder(nodeCount).
		GettingBlocksVia(mockBlockUtils).
		Build()

	return testNet
}

func CreateTestNetworkBuilder(nodeCount int) *TestNetworkBuilder {
	return &TestNetworkBuilder{
		logger:          &lh.SilentLogger{},
		customNodes:     nil,
		electionTrigger: nil,
		blockUtils:      nil,
		blocksPool:      nil,
		nodeCount:       nodeCount,
	}
}

func createNodes(nodeCount int, discovery gossip.Discovery) []*Node {
	nodes := make([]*Node, nodeCount)

	for i := range nodes {
		nodes[i] = buildNode(lh.PublicKey(fmt.Sprintf("Node %d", i)), discovery)
	}

	// TODO postpone handling custom nodes till needed by TDD (see TestNetworkBuilder.ts)

	return nodes
}

func (net *TestNetwork) GetNodeGossip(pk lh.PublicKey) (*gossip.Gossip, bool) {
	return net.discovery.GetGossipByPK(pk)
}

func (net *TestNetwork) Start() {

}

func (net *TestNetwork) Stop() {

}
