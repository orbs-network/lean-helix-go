package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type NodeBuilder struct {
	networkCommunication lh.NetworkCommunication
	publicKey            Ed25519PublicKey
	storage              lh.Storage
	logger               log.BasicLogger
	electionTrigger      lh.ElectionTrigger
	blockUtils           lh.BlockUtils
	filter               *lh.NetworkMessageFilter
	logsToConsole        bool
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

func (builder *NodeBuilder) ThatIsPartOf(networkCommunication lh.NetworkCommunication) *NodeBuilder {
	if builder.networkCommunication == nil {
		builder.networkCommunication = networkCommunication
	}
	return builder
}

func (builder *NodeBuilder) ElectingLeaderUsing(electionTrigger lh.ElectionTrigger) *NodeBuilder {
	if builder.electionTrigger == nil {
		builder.electionTrigger = electionTrigger
	}
	return builder
}

func (builder *NodeBuilder) WithPublicKey(publicKey Ed25519PublicKey) *NodeBuilder {
	if builder.publicKey == nil {
		builder.publicKey = publicKey
	}
	return builder
}

func (builder *NodeBuilder) buildConfig() *lh.Config {
	// TODO consider using members of node builder in the config if they are defined, like in TS code (NodeBuilder.ts)

	var (
		electionTrigger lh.ElectionTrigger
		blockUtils      lh.BlockUtils
		storage         lh.Storage
		keyManager      lh.KeyManager
	)

	if builder.electionTrigger != nil {
		electionTrigger = builder.electionTrigger
	} else {
		electionTrigger = NewMockElectionTrigger()
	}

	if builder.blockUtils != nil {
		blockUtils = builder.blockUtils
	} else {
		blockUtils = NewMockBlockUtils(nil)
	}

	if builder.storage != nil {
		storage = builder.storage
	} else {
		storage = lh.NewInMemoryStorage()
	}

	publicKey := builder.publicKey
	if publicKey == nil {
		publicKey = Ed25519PublicKey(fmt.Sprintf("Dummy PublicKey"))
	}
	keyManager = NewMockKeyManager(publicKey)

	return &lh.Config{
		NetworkCommunication: builder.networkCommunication,
		ElectionTrigger:      electionTrigger,
		BlockUtils:           blockUtils,
		KeyManager:           keyManager,
		Logger:               builder.logger.For(log.Service("node")),
		Storage:              storage,
	}
}

func (builder *NodeBuilder) GettingBlocksVia(blockUtils lh.BlockUtils) *NodeBuilder {
	if builder.blockUtils == nil {
		builder.blockUtils = blockUtils
	}
	return builder
}

func (builder *NodeBuilder) ThatLogsTo(logger log.BasicLogger) *NodeBuilder {
	if builder.logger == nil {
		builder.logger = logger
	}
	return builder
}

func (builder *NodeBuilder) Build() *Node {
	nodeConfig := builder.buildConfig()
	leanHelix := lh.NewLeanHelix(nodeConfig)
	node := &Node{
		Config:     nodeConfig,
		leanHelix:  leanHelix,
		blockChain: NewInMemoryBlockChain(),
	}
	leanHelix.RegisterOnCommitted(node.onCommittedBlock)
	return node
}

func buildNode(
	publicKey Ed25519PublicKey,
	discovery gossip.Discovery,
	logger log.BasicLogger) *Node {

	nodeLogger := logger.For(log.Service("node"))
	electionTrigger := NewMockElectionTrigger()
	blockUtils := NewMockBlockUtils(nil)
	gossip := gossip.NewGossip(discovery, publicKey)
	discovery.RegisterGossip(publicKey, gossip)

	return NewNodeBuilder().
		ThatIsPartOf(gossip).
		GettingBlocksVia(blockUtils).
		ElectingLeaderUsing(electionTrigger).
		WithPublicKey(publicKey).
		ThatLogsTo(nodeLogger).
		Build()
}
