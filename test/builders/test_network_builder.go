package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetworkBuilder struct {
	nodeCount            int
	electionTrigger      lh.ElectionTrigger
	blockUtils           *MockBlockUtils
	blocksPool           []lh.Block
	nonMemberNodeIndices []int
	discovery            gossip.Discovery
	nodesBlockHeight     BlockHeight
}

func (builder *TestNetworkBuilder) RequestBlocksWith(utils *MockBlockUtils) *TestNetworkBuilder {
	builder.blockUtils = utils
	return builder
}

func (builder *TestNetworkBuilder) WithNodeCount(nodeCount int) *TestNetworkBuilder {
	builder.nodeCount = nodeCount
	return builder
}

func (builder *TestNetworkBuilder) Build() *TestNetwork {

	return &TestNetwork{
		Nodes:      builder.CreateNodes(),
		BlockUtils: builder.blockUtils,
		Discovery:  builder.discovery,
	}

	// TODO Why we need this?? it does nothing on TS code
	//testNet.registerNodes()
	//return testNet
}

func NewSimpleTestNetwork(
	nodeCount int,
	nodesBlockHeight BlockHeight,
	blocksPool []lh.Block) *TestNetwork {

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

	return NewTestNetworkBuilder(nodeCount).
		WithBlockHeight(nodesBlockHeight).
		RequestBlocksWith(mockBlockUtils).
		Build()
}

func NewTestNetworkBuilder(nodeCount int) *TestNetworkBuilder {
	return &TestNetworkBuilder{
		nodeCount:       nodeCount,
		electionTrigger: nil,
		blockUtils:      nil,
		blocksPool:      nil,
		discovery:       gossip.NewGossipDiscovery(),
	}
}

func (builder *TestNetworkBuilder) CreateNodes() []*Node {
	nodes := make([]*Node, builder.nodeCount)

	for i := range nodes {
		nodes[i] = buildNode(Ed25519PublicKey(fmt.Sprintf("Node %d", i)), builder.discovery)
	}
	for _, idx := range builder.nonMemberNodeIndices {
		builder.discovery.UnregisterGossip(nodes[idx].Config.KeyManager.MyPublicKey())
	}
	return nodes
}

func (builder *TestNetworkBuilder) WithBlockHeight(height BlockHeight) *TestNetworkBuilder {
	builder.nodesBlockHeight = height
	return builder
}
