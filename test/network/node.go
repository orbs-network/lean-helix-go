package network

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

type NodeState struct {
	block           interfaces.Block
	validationCount int
}

type Node struct {
	leanHelix        *leanhelix.LeanHelix
	blockChain       *mocks.InMemoryBlockChain
	ElectionTrigger  *mocks.ElectionTriggerMock
	BlockUtils       *mocks.MockBlockUtils
	KeyManager       *mocks.MockKeyManager
	Storage          interfaces.Storage
	Gossip           *mocks.CommunicationMock
	Membership       interfaces.Membership
	MemberId         primitives.MemberId
	NodeStateChannel chan *NodeState
}

func (node *Node) GetKeyManager() interfaces.KeyManager {
	return node.KeyManager
}

func (node *Node) GetMemberId() primitives.MemberId {
	return node.MemberId
}

func (node *Node) GetLatestBlock() interfaces.Block {
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

func (node *Node) onCommittedBlock(ctx context.Context, block interfaces.Block, blockProof []byte) {
	node.blockChain.AppendBlockToChain(block, blockProof)
	node.NodeStateChannel <- &NodeState{
		block:           block,
		validationCount: node.BlockUtils.ValidationCounter,
	}
}

func (node *Node) StartConsensus(ctx context.Context) {
	if node.leanHelix != nil {
		go node.leanHelix.Run(ctx)
		node.leanHelix.UpdateState(ctx, node.GetLatestBlock(), nil)
	}
}

func (node *Node) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProof []byte) bool {
	return node.leanHelix.ValidateBlockConsensus(ctx, block, blockProof)
}

func (node *Node) Sync(ctx context.Context, prevBlock interfaces.Block, blockProof []byte) {
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

func (node *Node) BuildConfig(logger interfaces.Logger) *interfaces.Config {
	return &interfaces.Config{
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
	membership interfaces.Membership,
	gossip *mocks.CommunicationMock,
	blockUtils *mocks.MockBlockUtils,
	electionTrigger *mocks.ElectionTriggerMock,
	logger interfaces.Logger) *Node {

	memberId := membership.MyMemberId()
	node := &Node{
		blockChain:       mocks.NewInMemoryBlockChain(),
		ElectionTrigger:  electionTrigger,
		BlockUtils:       blockUtils,
		KeyManager:       mocks.NewMockKeyManager(memberId),
		Storage:          storage.NewInMemoryStorage(),
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
