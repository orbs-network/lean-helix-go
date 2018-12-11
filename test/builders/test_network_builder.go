package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetworkBuilder struct {
	NodeCount            int
	customNodeBuilders   []*NodeBuilder
	upcomingBlocks       []lh.Block
	keyManager           lh.KeyManager
	blockUtils           lh.BlockUtils
	networkCommunication lh.NetworkCommunication
}

func (builder *TestNetworkBuilder) WithNodeCount(nodeCount int) *TestNetworkBuilder {
	builder.NodeCount = nodeCount
	return builder
}

func (builder *TestNetworkBuilder) WithCustomNodeBuilder(nodeBuilder *NodeBuilder) *TestNetworkBuilder {
	builder.customNodeBuilders = append(builder.customNodeBuilders, nodeBuilder)
	return builder
}

func (builder *TestNetworkBuilder) WithBlocks(upcomingBlocks []lh.Block) *TestNetworkBuilder {
	if builder.upcomingBlocks == nil {
		builder.upcomingBlocks = upcomingBlocks
	}
	return builder
}

func (builder *TestNetworkBuilder) Build() *TestNetwork {
	blocksPool := builder.buildBlocksPool()
	discovery := gossip.NewGossipDiscovery()
	nodes := builder.createNodes(discovery, blocksPool)
	testNetwork := NewTestNetwork(discovery)
	testNetwork.RegisterNodes(nodes)
	return testNetwork
}

func (builder *TestNetworkBuilder) buildBlocksPool() *BlocksPool {
	if builder.upcomingBlocks == nil {
		b1 := CreateBlock(GenesisBlock)
		b2 := CreateBlock(b1)
		b3 := CreateBlock(b2)
		b4 := CreateBlock(b3)

		return NewBlocksPool([]lh.Block{b1, b2, b3, b4})
	} else {
		return NewBlocksPool(builder.upcomingBlocks)
	}
}

func (builder *TestNetworkBuilder) buildNode(
	nodeBuilder *NodeBuilder,
	publicKey primitives.Ed25519PublicKey,
	discovery *gossip.Discovery,
	blocksPool *BlocksPool) *Node {

	gossip := gossip.NewGossip(discovery)
	discovery.RegisterGossip(publicKey, gossip)
	return nodeBuilder.ThatIsPartOf(gossip).WithBlocksPool(blocksPool).WithPublicKey(publicKey).Build()
}

func (builder *TestNetworkBuilder) createNodes(discovery *gossip.Discovery, blocksPool *BlocksPool) []*Node {
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

func (builder *TestNetworkBuilder) WithNetworkCommunication(comm lh.NetworkCommunication) *TestNetworkBuilder {
	builder.networkCommunication = comm
	return builder
}

func (builder *TestNetworkBuilder) WithBlockUtils(utils lh.BlockUtils) *TestNetworkBuilder {
	builder.blockUtils = utils
	return builder
}

func (builder *TestNetworkBuilder) WithKeyManager(mgr lh.KeyManager) *TestNetworkBuilder {
	builder.keyManager = mgr
	return builder
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

func ATestNetwork(countOfNodes int, blocksPool ...lh.Block) *TestNetwork {
	return NewTestNetworkBuilder().WithNodeCount(countOfNodes).WithBlocks(blocksPool).Build()
}

func CreateTestNetworkForConsumerTests(
	countOfNodes int,
	spi *lh.LeanHelixSPI,
	blocks []lh.Block,
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
