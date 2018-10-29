package builders

import (
	"context"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
)

// The Lean Helix BlockHeight component tests in this file require a test network to run

const MINIMUM_NODES = 2

type TestNetwork struct {
	mock.Mock
	ctx        context.Context
	ctxCancel  context.CancelFunc
	Nodes      []*Node
	BlockUtils *MockBlockUtils
	Transport  *MockNetworkCommunication
	Discovery  gossip.Discovery
}

type TestNetworkBuilder struct {
	nodeCount            int
	electionTrigger      lh.ElectionTrigger
	blockUtils           *MockBlockUtils
	blocksPool           []lh.Block
	nonMemberNodeIndices []int
	discovery            gossip.Discovery
	logger               log.BasicLogger
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

func (builder *TestNetworkBuilder) ThatLogsToCustomLogger(logger log.BasicLogger) *TestNetworkBuilder {
	builder.logger = logger
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

	//ctx, ctxCancel := context.WithCancel(context.Background())

	var output io.Writer
	output = os.Stdout

	testId := log.RandNumStr()
	testLogger := log.GetLogger(log.String("test-id", testId)).
		WithOutput(log.NewOutput(output).
			WithFormatter(log.NewHumanReadableFormatter()))
	//WithFilter(log.String("flow", "MockBlock-sync")).
	//WithFilter(log.String("service", "gossip"))
	testLogger.Info("===========================================================================")

	return &TestNetworkBuilder{
		nodeCount:       nodeCount,
		electionTrigger: nil,
		blockUtils:      nil,
		blocksPool:      nil,
		discovery:       gossip.NewGossipDiscovery(),
		logger:          testLogger,
	}
}

func (builder *TestNetworkBuilder) CreateNodes() []*Node {
	nodes := make([]*Node, builder.nodeCount)

	for i := range nodes {
		nodes[i] = buildNode(Ed25519PublicKey(fmt.Sprintf("Node %d", i)), builder.discovery, builder.logger)
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

func (net *TestNetwork) GetNodeGossip(pk Ed25519PublicKey) *gossip.Gossip {
	return net.Discovery.GetGossipByPK(pk)
}

func (net *TestNetwork) TriggerElection(ctx context.Context) {
	for _, node := range net.Nodes {
		node.TriggerElection(ctx)
	}
}

func (net *TestNetwork) StartConsensusOnAllNodes() error {
	if len(net.Nodes) < MINIMUM_NODES {
		return fmt.Errorf("not enough nodes in test network - found %d but minimum is %d", len(net.Nodes), MINIMUM_NODES)
	}
	for _, node := range net.Nodes {
		node.StartConsensus()
	}
	return nil
}

func (net *TestNetwork) Stop() {

	net.ctxCancel()

	// TODO Do we need this??
	for _, node := range net.Nodes {
		node.Dispose()
	}

}
