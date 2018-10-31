package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetworkBuilder struct {
	NodeCount          int
	customNodeBuilders []*NodeBuilder
	blocksPool         []lh.Block
}

func (builder *TestNetworkBuilder) WithNodeCount(nodeCount int) *TestNetworkBuilder {
	builder.NodeCount = nodeCount
	return builder
}

func (builder *TestNetworkBuilder) WithCustomNodeBuilder(nodeBuilder *NodeBuilder) *TestNetworkBuilder {
	builder.customNodeBuilders = append(builder.customNodeBuilders, nodeBuilder)
	return builder
}

func (builder *TestNetworkBuilder) WithBlocksPool(blocksPool []lh.Block) *TestNetworkBuilder {
	if builder.blocksPool == nil {
		builder.blocksPool = blocksPool
	}
	return builder
}

func (builder *TestNetworkBuilder) Build() *TestNetwork {
	blocksPool := builder.buildBlocksPool()
	discovery := gossip.NewGossipDiscovery()
	nodes := builder.createNodes(discovery, blocksPool)
	testNetwork := NewTestNetwork(discovery, blocksPool)
	testNetwork.RegisterNodes(nodes)
	return testNetwork
}

func (builder *TestNetworkBuilder) buildBlocksPool() []lh.Block {
	if builder.blocksPool == nil {
		b1 := CreateBlock(GenesisBlock)
		b2 := CreateBlock(b1)
		b3 := CreateBlock(b2)
		b4 := CreateBlock(b3)

		return []lh.Block{b1, b2, b3, b4}
	} else {
		return builder.blocksPool
	}
}

func (builder *TestNetworkBuilder) buildNode(
	nodeBuilder *NodeBuilder,
	publicKey primitives.Ed25519PublicKey,
	discovery *gossip.Discovery,
	blocksPool []lh.Block) *Node {

	gossip := gossip.NewGossip(discovery)
	discovery.RegisterGossip(publicKey, gossip)
	return nodeBuilder.ThatIsPartOf(gossip).WithBlocksPool(blocksPool).WithPublicKey(publicKey).Build()
}

func (builder *TestNetworkBuilder) createNodes(discovery *gossip.Discovery, blocksPool []lh.Block) []*Node {
	var nodes []*Node
	for i := 0; i < builder.NodeCount; i++ {
		nodeBuilder := NewNodeBuilder()
		publicKey := primitives.Ed25519PublicKey(fmt.Sprintf("Node %d", i))
		node := builder.buildNode(nodeBuilder, publicKey, discovery, blocksPool)
		nodes = append(nodes, node)
	}

	for i, customBuilder := range builder.customNodeBuilders {
		publicKey := primitives.Ed25519PublicKey(fmt.Sprintf("Custom-Node %d", i))
		node := builder.buildNode(customBuilder, publicKey, discovery, blocksPool)
		nodes = append(nodes, node)
	}

	return nodes
}

func NewTestNetworkBuilder() *TestNetworkBuilder {
	return &TestNetworkBuilder{
		NodeCount:          0,
		customNodeBuilders: nil,
		blocksPool:         nil,
	}
}

func ATestNetwork(countOfNodes int, blocksPool []lh.Block) *TestNetwork {
	testNetwork := NewTestNetworkBuilder()
	return testNetwork.WithNodeCount(countOfNodes).WithBlocksPool(blocksPool).Build()
}
