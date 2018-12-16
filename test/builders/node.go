package builders

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type NodeState struct {
	block           lh.Block
	validationCount int
}

type Node struct {
	leanHelix        *lh.LeanHelix
	blockChain       *InMemoryBlockChain
	ElectionTrigger  *ElectionTriggerMock
	BlockUtils       *MockBlockUtils
	KeyManager       *MockKeyManager
	Storage          lh.Storage
	Gossip           *gossip.Gossip
	PublicKey        primitives.MemberId
	NodeStateChannel chan *NodeState
}

func (node *Node) GetLatestBlock() lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) TriggerElection() {
	node.ElectionTrigger.ManualTrigger()
}

func (node *Node) TriggerElectionSync(ctx context.Context) {
	node.ElectionTrigger.ManualTriggerSync(ctx)
}

func (node *Node) onCommittedBlock(block lh.Block) {
	node.blockChain.AppendBlockToChain(block)
	node.NodeStateChannel <- &NodeState{
		block:           block,
		validationCount: node.BlockUtils.validationCounter,
	}
}

func (node *Node) StartConsensus(ctx context.Context) {
	if node.leanHelix != nil {
		go node.leanHelix.Run(ctx)
		node.leanHelix.UpdateConsensusRound(node.GetLatestBlock())
	}
}

func (node *Node) IsLeader() bool {
	return node.leanHelix != nil && node.leanHelix.IsLeader()
}

func (node *Node) Tick(ctx context.Context) {
	node.leanHelix.Tick(ctx)
}
func (node *Node) StartConsensusSync() {
	if node.leanHelix != nil {
		go node.leanHelix.UpdateConsensusRound(node.GetLatestBlock())
	}
}

func (node *Node) BuildConfig(logger lh.Logger) *lh.Config {
	return &lh.Config{
		NetworkCommunication: node.Gossip,
		ElectionTrigger:      node.ElectionTrigger,
		BlockUtils:           node.BlockUtils,
		KeyManager:           node.KeyManager,
		Storage:              node.Storage,
		Logger:               logger,
	}

}

func NewNode(
	publicKey primitives.MemberId,
	gossip *gossip.Gossip,
	blockUtils *MockBlockUtils,
	electionTrigger *ElectionTriggerMock,
	logger lh.Logger) *Node {
	node := &Node{
		blockChain:       NewInMemoryBlockChain(),
		ElectionTrigger:  electionTrigger,
		BlockUtils:       blockUtils,
		KeyManager:       NewMockKeyManager(publicKey),
		Storage:          lh.NewInMemoryStorage(),
		Gossip:           gossip,
		PublicKey:        publicKey,
		NodeStateChannel: make(chan *NodeState),
	}

	leanHelix := lh.NewLeanHelix(node.BuildConfig(logger))
	leanHelix.RegisterOnCommitted(node.onCommittedBlock)
	gossip.RegisterOnMessage(leanHelix.GossipMessageReceived)

	node.leanHelix = leanHelix
	return node

}
