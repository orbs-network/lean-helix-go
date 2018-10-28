package builders

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type NodeBuilder struct {
	ctx             context.Context
	ctxCancel       context.CancelFunc
	gossip          *gossip.Gossip
	keyManager      *mockKeyManager
	storage         lh.Storage
	logger          log.BasicLogger
	electionTrigger lh.ElectionTrigger
	blockUtils      lh.BlockUtils
	filter          *lh.NetworkMessageFilter
	logsToConsole   bool
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{
		gossip:          nil,
		keyManager:      nil,
		storage:         nil,
		logger:          nil,
		electionTrigger: nil,
		blockUtils:      nil,
		filter:          nil,
		logsToConsole:   false,
	}
}

func (builder *NodeBuilder) ThatIsPartOf(gossip *gossip.Gossip) *NodeBuilder {
	if builder.gossip == nil {
		builder.gossip = gossip
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
	if builder.keyManager == nil {
		builder.keyManager = NewMockKeyManager(publicKey)
	}
	return builder
}

func (builder *NodeBuilder) buildConfig() *lh.Config {
	// TODO consider using members of node builder in the config if they are defined, like in TS code (NodeBuilder.ts)

	var (
		electionTrigger lh.ElectionTrigger
		blockUtils      lh.BlockUtils
		storage         lh.Storage
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

	//if builder.logger != nil {
	//	logger = builder.logger
	//} else {
	//	if builder.logsToConsole {
	//		// TODO Find the silent logger
	//		//logger = lh.NewConsoleLogger(string(builder.publicKey))
	//		logger =
	//	} else {
	//		logger = log.BasicLogger.For(log.Service("node"))
	//	}
	//}

	if builder.storage != nil {
		storage = builder.storage
	} else {
		storage = lh.NewInMemoryStorage()
	}

	return &lh.Config{
		NetworkCommunication: builder.gossip,
		ElectionTrigger:      electionTrigger,
		BlockUtils:           blockUtils,
		KeyManager:           builder.keyManager,
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

	//func NewNode(ctx context.Context, ctxCancel context.CancelFunc, keyManager lh.KeyManager, config *lh.Config) *Node {
	nodeConfig := builder.buildConfig()
	leanHelix := lh.NewLeanHelix(builder.ctx, builder.ctxCancel, nodeConfig)
	node := &Node{
		KeyManager: builder.keyManager,
		Filter:     builder.filter,
		Gossip:     builder.gossip,
		Config:     nodeConfig,
		leanHelix:  leanHelix,
		blockChain: NewInMemoryBlockChain(),
	}
	leanHelix.RegisterOnCommitted(node.onCommittedBlock)
	return node
}

func (builder *NodeBuilder) WithContext(ctx context.Context, ctxCancel context.CancelFunc) *NodeBuilder {
	builder.ctx = ctx
	builder.ctxCancel = ctxCancel
	return builder
}

func (builder *NodeBuilder) WithFilter(filter *lh.NetworkMessageFilter) *NodeBuilder {
	builder.filter = filter
	return builder
}
