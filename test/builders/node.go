package builders

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type NodeState struct {
	block           leanhelix.Block
	validationCount int
}

type Node struct {
	leanHelix        *leanhelix.LeanHelix
	blockChain       *InMemoryBlockChain
	ElectionTrigger  *ElectionTriggerMock
	BlockUtils       *MockBlockUtils
	KeyManager       *MockKeyManager
	Storage          leanhelix.Storage
	Gossip           *gossip.Gossip
	Membership       leanhelix.Membership
	MemberId         primitives.MemberId
	NodeStateChannel chan *NodeState
}

func (node *Node) GetLatestBlock() leanhelix.Block {
	return node.blockChain.GetLastBlock()
}

func (node *Node) GetLatestBlockProof() []byte {
	return node.blockChain.GetLastBlockProof()
}

func (node *Node) TriggerElection() {
	node.ElectionTrigger.ManualTrigger()
}

func (node *Node) TriggerElectionSync(ctx context.Context) {
	node.ElectionTrigger.ManualTriggerSync(ctx)
}

func (node *Node) onCommittedBlock(ctx context.Context, block leanhelix.Block, blockProof []byte) {
	node.blockChain.AppendBlockToChain(block, blockProof)
	node.NodeStateChannel <- &NodeState{
		block:           block,
		validationCount: node.BlockUtils.validationCounter,
	}
}

func (node *Node) StartConsensus(ctx context.Context) {
	if node.leanHelix != nil {
		go node.leanHelix.Run(ctx)
		node.leanHelix.UpdateState(ctx, node.GetLatestBlock(), nil)
	}
}

func (node *Node) ValidateBlockConsensus(ctx context.Context, block leanhelix.Block, blockProof []byte) bool {
	return node.leanHelix.ValidateBlockConsensus(ctx, block, blockProof)
}

func (node *Node) Sync(ctx context.Context, prevBlock leanhelix.Block, blockProof []byte) {
	if node.leanHelix != nil {
		if node.leanHelix.ValidateBlockConsensus(ctx, prevBlock, blockProof) {
			go node.leanHelix.UpdateState(ctx, prevBlock, nil)
		}
	}
}

func (node *Node) Tick(ctx context.Context) {
	node.leanHelix.Tick(ctx)
}
func (node *Node) StartConsensusSync(ctx context.Context) {
	if node.leanHelix != nil {
		go node.leanHelix.UpdateState(ctx, node.GetLatestBlock(), nil)
	}
}

func (node *Node) BuildConfig(logger leanhelix.Logger) *leanhelix.Config {
	return &leanhelix.Config{
		Communication:   node.Gossip,
		Membership:      node.Membership,
		ElectionTrigger: node.ElectionTrigger,
		BlockUtils:      node.BlockUtils,
		KeyManager:      node.KeyManager,
		Storage:         node.Storage,
		Logger:          logger,
	}

}

func NewNode(
	membership leanhelix.Membership,
	gossip *gossip.Gossip,
	blockUtils *MockBlockUtils,
	electionTrigger *ElectionTriggerMock,
	logger leanhelix.Logger) *Node {

	memberId := membership.MyMemberId()
	node := &Node{
		blockChain:       NewInMemoryBlockChain(),
		ElectionTrigger:  electionTrigger,
		BlockUtils:       blockUtils,
		KeyManager:       NewMockKeyManager(memberId),
		Storage:          leanhelix.NewInMemoryStorage(),
		Gossip:           gossip,
		Membership:       membership,
		MemberId:         memberId,
		NodeStateChannel: make(chan *NodeState),
	}

	leanHelix := leanhelix.NewLeanHelix(node.BuildConfig(logger), node.onCommittedBlock)
	gossip.RegisterOnMessage(leanHelix.HandleConsensusMessage)

	node.leanHelix = leanHelix
	return node

}
