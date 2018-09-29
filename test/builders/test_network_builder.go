package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/mock"
)

const MINIMUM_NODES = 2

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
	blockUtils      *MockBlockUtils
	blocksPool      []lh.Block
	nodeCount       int
	discovery       gossip.Discovery
}

func (builder *TestNetworkBuilder) GettingBlocksVia(utils *MockBlockUtils) *TestNetworkBuilder {
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

	testNet := &TestNetwork{

		Nodes:      builder.createNodes(),
		BlockUtils: builder.blockUtils,
		//Transport:  builder.transport,
		discovery: builder.discovery,
	}

	// TODO Why we need this?? it does nothing on TS code
	//testNet.registerNodes()

	return testNet
}

func NewSimpleTestNetwork(nodeCount int, blocksPool []lh.Block) *TestNetwork {

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

	testNet := NewTestNetworkBuilder(nodeCount).
		GettingBlocksVia(mockBlockUtils).
		Build()

	return testNet
}

func NewTestNetworkBuilder(nodeCount int) *TestNetworkBuilder {
	return &TestNetworkBuilder{
		logger:          &lh.SilentLogger{},
		customNodes:     nil,
		electionTrigger: nil,
		blockUtils:      nil,
		blocksPool:      nil,
		nodeCount:       nodeCount,
		discovery:       gossip.NewGossipDiscovery(),
	}
}

func (builder *TestNetworkBuilder) createNodes() []*Node {
	nodes := make([]*Node, builder.nodeCount)

	for i := range nodes {
		nodes[i] = buildNode(lh.PublicKey(fmt.Sprintf("Node %d", i)), builder.discovery)
	}

	// TODO postpone handling custom nodes till needed by TDD (see TestNetworkBuilder.ts)

	return nodes
}

func (net *TestNetwork) GetNodeGossip(pk lh.PublicKey) (*gossip.Gossip, bool) {
	return net.discovery.GetGossipByPK(pk)
}

func (net *TestNetwork) StartConsensusOnAllNodes() error {
	if len(net.Nodes) < 2 {
		return fmt.Errorf("not enough nodes in test network - found %d but minimum is %d", len(net.Nodes), MINIMUM_NODES)
	}
	for _, node := range net.Nodes {
		node.StartConsensus()
	}
	return nil
}

func (net *TestNetwork) Stop() {
	for _, node := range net.Nodes {
		node.Dispose()
	}

}
