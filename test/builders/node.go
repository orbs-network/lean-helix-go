package builders

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type Node struct {
	KeyManager lh.KeyManager
	Config     *lh.Config
	leanHelix  lh.LeanHelix
	blockChain *InMemoryBlockChain
	Gossip     *gossip.Gossip
}

func NewNode(ctx context.Context, ctxCancel context.CancelFunc, keyManager lh.KeyManager, config *lh.Config) *Node {
	pbft := lh.NewLeanHelix(ctx, ctxCancel, config)
	node := &Node{
		KeyManager: keyManager,
		Config:     config,
		leanHelix:  pbft,
		blockChain: NewInMemoryBlockChain(),
	}
	pbft.RegisterOnCommitted(node.onCommittedBlock)
	return node
}

func buildNode(ctx context.Context, ctxCancel context.CancelFunc, publicKey Ed25519PublicKey, discovery gossip.Discovery, logger log.BasicLogger) *Node {

	nodeLogger := logger.For(log.Service("node"))
	electionTrigger := NewMockElectionTrigger() // TODO TestNetworkBuilder.ts uses ElectionTriggerFactory here, maybe do it too
	blockUtils := NewMockBlockUtils(nil)
	gossip := gossip.NewGossip(discovery, publicKey)
	discovery.RegisterGossip(publicKey, gossip)

	return NewNodeBuilder().
		WithContext(ctx, ctxCancel).
		ThatIsPartOf(gossip).
		GettingBlocksVia(blockUtils).
		ElectingLeaderUsing(electionTrigger).
		WithPublicKey(publicKey).
		ThatLogsTo(nodeLogger).
		Build()
}

func (node *Node) GetLatestCommittedBlock() lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) IsLeader() bool {
	return node.leanHelix.IsLeader()
}

func (node *Node) TriggerElection() {
	node.Config.ElectionTrigger.(*ElectionTriggerMock).Trigger()
}

func (node *Node) onCommittedBlock(block lh.Block) {
	node.blockChain.AppendBlockToChain(block)
}

func (node *Node) StartConsensus() {
	if node.leanHelix != nil {
		lastCommittedBlock := node.GetLatestCommittedBlock()
		node.leanHelix.Start(lastCommittedBlock.Height())
	}
}

func (node *Node) Dispose() {
	if node.leanHelix != nil {
		node.leanHelix.Dispose()
	}
}
