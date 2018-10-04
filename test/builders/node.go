package builders

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type Node struct {
	PublicKey  lh.PublicKey
	Config     *lh.Config
	leanHelix  lh.LeanHelix
	blockChain *InMemoryBlockChain
	Gossip     *gossip.Gossip
}

func NewNode(publicKey lh.PublicKey, config *lh.Config) *Node {
	pbft := lh.NewLeanHelix(config)
	node := &Node{
		PublicKey:  publicKey,
		Config:     config,
		leanHelix:  pbft,
		blockChain: NewInMemoryBlockChain(),
	}
	pbft.RegisterOnCommitted(node.onCommittedBlock)
	return node
}

func buildNode(ctx context.Context, publicKey lh.PublicKey, discovery gossip.Discovery, logger log.BasicLogger) *Node {

	nodeLogger := logger.For(log.Service("node"))
	electionTrigger := NewMockElectionTrigger() // TODO TestNetworkBuilder.ts uses ElectionTriggerFactory here, maybe do it too
	blockUtils := NewMockBlockUtils(nil)
	gossip := gossip.NewGossip(discovery)
	discovery.RegisterGossip(publicKey, gossip)
	networkCommunication := NewInMemoryNetworkCommunication(discovery, gossip)

	return NewNodeBuilder().
		WithContext(ctx).
		ThatIsPartOf(networkCommunication).
		GettingBlocksVia(blockUtils).
		ElectingLeaderUsing(electionTrigger).
		WithPK(publicKey).
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
		node.leanHelix.Start(lastCommittedBlock.GetHeight())
	}
}

func (node *Node) Dispose() {
	if node.leanHelix != nil {
		node.leanHelix.Dispose()
	}
}
