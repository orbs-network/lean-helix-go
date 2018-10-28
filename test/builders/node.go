package builders

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type Node struct {
	KeyManager *mockKeyManager
	Config     *lh.Config
	leanHelix  lh.LeanHelix
	blockChain *InMemoryBlockChain
	Gossip     *gossip.Gossip
	Filter     *lh.NetworkMessageFilter
}

func buildNode(
	publicKey Ed25519PublicKey,
	nodeBlockHeight BlockHeight,
	discovery gossip.Discovery,
	logger log.BasicLogger) *Node {

	nodeLogger := logger.For(log.Service("node"))
	electionTrigger := NewMockElectionTrigger() // TODO TestNetworkBuilder.ts uses ElectionTriggerFactory here, maybe do it too
	blockUtils := NewMockBlockUtils(nil)
	gossip := gossip.NewGossip(discovery, publicKey)
	discovery.RegisterGossip(publicKey, gossip)
	mockReceiver := NewMockMessageReceiver()
	filter := lh.NewNetworkMessageFilter(gossip, publicKey, mockReceiver)
	ctx, _ := context.WithCancel(context.Background())
	filter.SetBlockHeight(ctx, nodeBlockHeight)

	return NewNodeBuilder().
		WithFilter(filter).
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

func (node *Node) TriggerElection(ctx context.Context) {
	node.Config.ElectionTrigger.(*ElectionTriggerMock).Trigger(ctx)
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
