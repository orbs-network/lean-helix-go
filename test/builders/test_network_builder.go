package builders

import (
	"context"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/orbs-network/orbs-network-go/config"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
)

// The Lean Helix BlockHeight component tests in this file require a test network to run

const MINIMUM_NODES = 2

type TestNetwork struct {
	mock.Mock
	ctxCancel  context.CancelFunc
	Nodes      []*Node
	BlockUtils *MockBlockUtils
	Transport  *MockNetworkCommunication
	discovery  gossip.Discovery
}

type TestNetworkBuilder struct {
	ctx             context.Context
	ctxCancel       context.CancelFunc
	logger          log.BasicLogger
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

func (builder *TestNetworkBuilder) ThatLogsToCustomLogger(logger log.BasicLogger) *TestNetworkBuilder {
	builder.logger = logger
	return builder
}

func (builder *TestNetworkBuilder) Build() *TestNetwork {

	testNet := &TestNetwork{
		ctxCancel:  builder.ctxCancel,
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

	ctx, ctxCancel := context.WithCancel(context.Background())

	var output io.Writer
	output = os.Stdout

	testId := log.RandNumStr()

	if os.Getenv("NO_LOG_STDOUT") == "true" {
		logFile, err := os.OpenFile(config.GetProjectSourceRootPath()+"/logs/acceptance/"+testId+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		output = logFile
	}
	testLogger := log.GetLogger(log.String("test-id", testId)).
		WithOutput(log.NewOutput(output).
			WithFormatter(log.NewHumanReadableFormatter()))
	//WithFilter(log.String("flow", "MockBlock-sync")).
	//WithFilter(log.String("service", "gossip"))
	testLogger.Info("===========================================================================")

	return &TestNetworkBuilder{
		ctx:             ctx,
		ctxCancel:       ctxCancel,
		logger:          testLogger,
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
		nodes[i] = buildNode(builder.ctx, lh.PublicKey(fmt.Sprintf("Node %d", i)), builder.discovery, builder.logger)
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

	net.ctxCancel()

	// TODO Do we need this??
	for _, node := range net.Nodes {
		node.Dispose()
	}

}
