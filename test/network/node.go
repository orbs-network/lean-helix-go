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
	instanceId       primitives.InstanceId
	leanHelix        *leanhelix.LeanHelix
	blockChain       *mocks.InMemoryBlockChain
	ElectionTrigger  *mocks.ElectionTriggerMock
	BlockUtils       *mocks.MockBlockUtils
	KeyManager       *mocks.MockKeyManager
	Storage          interfaces.Storage
	Communication    *mocks.CommunicationMock
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

func (node *Node) GetBlockProofAt(height primitives.BlockHeight) []byte {
	return node.blockChain.GetBlockProofAt(height)
}

func (node *Node) TriggerElection(ctx context.Context) {
	node.ElectionTrigger.ManualTrigger(ctx)
}

func (node *Node) TriggerElectionSync(ctx context.Context) {
	node.ElectionTrigger.ManualTriggerSync(ctx)
}

func (node *Node) onCommittedBlock(ctx context.Context, block interfaces.Block, blockProof []byte) {
	node.blockChain.AppendBlockToChain(block, blockProof)

	nodeState := &NodeState{
		block:           block,
		validationCount: node.BlockUtils.ValidationCounter,
	}

	select {
	case <-ctx.Done():
		return

	case node.NodeStateChannel <- nodeState:
		return
	}

}

func (node *Node) StartConsensus(ctx context.Context) {
	if node.leanHelix != nil {
		go node.leanHelix.Run(ctx)
		node.leanHelix.UpdateState(ctx, node.GetLatestBlock(), nil)
	}
}

func (node *Node) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProof []byte, prevBlockProof []byte) error {
	return node.leanHelix.ValidateBlockConsensus(ctx, block, blockProof, prevBlockProof)
}

func (node *Node) Sync(ctx context.Context, prevBlock interfaces.Block, blockProofBytes []byte, prevBlockProofBytes []byte) {
	if node.leanHelix != nil {
		if err := node.leanHelix.ValidateBlockConsensus(ctx, prevBlock, blockProofBytes, prevBlockProofBytes); err == nil {
			go node.leanHelix.UpdateState(ctx, prevBlock, prevBlockProofBytes)
		}
	}
}

func (node *Node) StartConsensusSync(ctx context.Context) {
	if node.leanHelix != nil {
		go node.leanHelix.UpdateState(ctx, node.GetLatestBlock(), nil)
	}
}

func (node *Node) BuildConfig(logger interfaces.Logger) *interfaces.Config {
	return &interfaces.Config{
		InstanceId:      node.instanceId,
		Communication:   node.Communication,
		Membership:      node.Membership,
		ElectionTrigger: node.ElectionTrigger,
		BlockUtils:      node.BlockUtils,
		KeyManager:      node.KeyManager,
		Storage:         node.Storage,
		Logger:          logger,
	}

}

func NewNode(
	instanceId primitives.InstanceId,
	membership interfaces.Membership,
	communication *mocks.CommunicationMock,
	blockUtils *mocks.MockBlockUtils,
	electionTrigger *mocks.ElectionTriggerMock,
	logger interfaces.Logger) *Node {

	memberId := membership.MyMemberId()
	node := &Node{
		instanceId:       instanceId,
		blockChain:       mocks.NewInMemoryBlockChain(),
		ElectionTrigger:  electionTrigger,
		BlockUtils:       blockUtils,
		KeyManager:       mocks.NewMockKeyManager(memberId),
		Storage:          storage.NewInMemoryStorage(),
		Communication:    communication,
		Membership:       membership,
		MemberId:         memberId,
		NodeStateChannel: make(chan *NodeState),
	}

	leanHelix := leanhelix.NewLeanHelix(node.BuildConfig(logger), node.onCommittedBlock)
	communication.RegisterOnMessage(leanHelix.HandleConsensusRawMessage)

	node.leanHelix = leanHelix
	return node

}
