package builders

import (
	"context"
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

func (builder *TestNetworkBuilder) Build(ctx context.Context) *TestNetwork {
	blocksPool := builder.buildBlocksPool()
	discovery := gossip.NewGossipDiscovery()
	nodes := builder.createNodes(ctx, discovery, blocksPool)
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
	ctx context.Context,
	nodeBuilder *NodeBuilder,
	publicKey primitives.Ed25519PublicKey,
	discovery *gossip.Discovery,
	blocksPool []lh.Block) *Node {

	gossip := gossip.NewGossip(discovery)
	discovery.RegisterGossip(publicKey, gossip)
	return nodeBuilder.ThatIsPartOf(gossip).WithBlocksPool(blocksPool).WithPublicKey(publicKey).Build()
}

func (builder *TestNetworkBuilder) createNodes(ctx context.Context, discovery *gossip.Discovery, blocksPool []lh.Block) []*Node {
	var nodes []*Node
	for i := 0; i < builder.NodeCount; i++ {
		nodeBuilder := NewNodeBuilder()
		publicKey := primitives.Ed25519PublicKey(fmt.Sprintf("Node %d", i))
		node := builder.buildNode(ctx, nodeBuilder, publicKey, discovery, blocksPool)
		nodes = append(nodes, node)
	}

	for i, customBuilder := range builder.customNodeBuilders {
		publicKey := primitives.Ed25519PublicKey(fmt.Sprintf("Custom-Node %d", i))
		node := builder.buildNode(ctx, customBuilder, publicKey, discovery, blocksPool)
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

func ABasicTestNetwork(ctx context.Context) *TestNetwork {
	return ATestNetwork(ctx, 4)
}

func ATestNetwork(ctx context.Context, countOfNodes int, blocksPool ...lh.Block) *TestNetwork {
	testNetwork := NewTestNetworkBuilder()
	return testNetwork.WithNodeCount(countOfNodes).WithBlocksPool(blocksPool).Build(ctx)
}
