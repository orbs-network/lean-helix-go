package builders

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type NodeState struct {
	block           lh.Block
	validationCount int
}

type Node struct {
	leanHelix        lh.LeanHelix
	blockChain       *InMemoryBlockChain
	ElectionTrigger  *ElectionTriggerMock
	BlockUtils       *MockBlockUtils
	KeyManager       lh.KeyManager
	Storage          lh.Storage
	Gossip           *gossip.Gossip
	PublicKey        primitives.Ed25519PublicKey
	NodeStateChannel chan *NodeState
}

func (node *Node) GetLatestBlock() lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) TriggerElection() {
	node.ElectionTrigger.ManualTrigger()
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
		lastCommittedBlock := node.GetLatestBlock()
		go node.leanHelix.Run(ctx)
		node.leanHelix.AcknowledgeBlockConsensus(lastCommittedBlock)
	}
}

func (node *Node) BuildConfig() *lh.Config {
	return &lh.Config{
		NetworkCommunication: node.Gossip,
		ElectionTrigger:      node.ElectionTrigger,
		BlockUtils:           node.BlockUtils,
		KeyManager:           node.KeyManager,
		Storage:              node.Storage,
	}

}

func NewNode(
	publicKey primitives.Ed25519PublicKey,
	gossip *gossip.Gossip,
	blockUtils *MockBlockUtils,
	electionTrigger *ElectionTriggerMock) *Node {
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

	leanHelix := lh.NewLeanHelix(node.BuildConfig())
	leanHelix.RegisterOnCommitted(node.onCommittedBlock)

	node.leanHelix = leanHelix
	return node

}
