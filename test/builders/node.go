package builders

import (
	"context"
	"fmt"
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
	electionTrigger  *ElectionTriggerMock
	BlockUtils       *MockBlockUtils
	KeyManager       lh.KeyManager
	Storage          lh.Storage
	Gossip           *gossip.Gossip
	PublicKey        primitives.Ed25519PublicKey
	NodeStateChannel chan *NodeState
}

func (node *Node) GetLatestCommittedBlock() lh.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) IsLeader() bool {
	return node.leanHelix.IsLeader()
}

func (node *Node) TriggerElection() {
	node.electionTrigger.Trigger()
}

func (node *Node) onCommittedBlock(block lh.Block) {
	fmt.Println("COMMIT on", node.PublicKey.KeyForMap(), "block", block.BlockHash().String())
	node.blockChain.AppendBlockToChain(block)
	node.NodeStateChannel <- &NodeState{
		block:           block,
		validationCount: node.BlockUtils.validationCounter,
	}
}

func (node *Node) StartConsensus() {
	if node.leanHelix != nil {
		lastCommittedBlock := node.GetLatestCommittedBlock()
		node.leanHelix.Start(context.Background(), lastCommittedBlock.Height()+1)
	}
}

func (node *Node) Dispose() {
	if node.leanHelix != nil {
		node.leanHelix.Dispose()
	}
}

func (node *Node) BuildConfig() *lh.Config {
	return &lh.Config{
		NetworkCommunication: node.Gossip,
		ElectionTrigger:      node.electionTrigger,
		BlockUtils:           node.BlockUtils,
		KeyManager:           node.KeyManager,
		Storage:              node.Storage,
	}

}

func NewNode(publicKey primitives.Ed25519PublicKey, gossip *gossip.Gossip, blockUtils *MockBlockUtils) *Node {
	node := &Node{
		blockChain:       NewInMemoryBlockChain(),
		electionTrigger:  NewMockElectionTrigger(),
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
