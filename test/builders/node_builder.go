package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
)

type NodeBuilder struct {
	networkCommunication lh.NetworkCommunication
	publicKey            lh.PublicKey
	storage              lh.Storage
	logger               log.BasicLogger
	electionTrigger      lh.ElectionTrigger
	blockUtils           lh.BlockUtils
	logsToConsole        bool
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{
		networkCommunication: nil,
		publicKey:            nil,
		storage:              nil,
		logger:               nil,
		electionTrigger:      nil,
		blockUtils:           nil,
		logsToConsole:        false,
	}
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

func (builder *NodeBuilder) WithPK(publicKey lh.PublicKey) *NodeBuilder {
	if builder.publicKey.Equals(lh.PublicKey("")) {
		builder.publicKey = publicKey
	}
	return builder
}

func (builder *NodeBuilder) buildConfig() *lh.Config {
	// TODO consider using members of node builder in the config if they are defined, like in TS code (NodeBuilder.ts)

	var (
		electionTrigger lh.ElectionTrigger
		blockUtils      lh.BlockUtils
		logger          log.BasicLogger
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
		NetworkCommunication: builder.networkCommunication,
		ElectionTrigger:      electionTrigger,
		BlockUtils:           blockUtils,
		KeyManager:           NewMockKeyManager(builder.publicKey),
		Logger:               logger.For(log.Service("node")),
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
	return NewNode(builder.publicKey, builder.buildConfig())
}
