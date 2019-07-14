// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package network

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

type NodeState struct {
	block           interfaces.Block
	validationCount int
}

type Node struct {
	instanceId          primitives.InstanceId
	leanHelix           *leanhelix.MainLoop
	blockChain          *mocks.InMemoryBlockChain
	ElectionTrigger     interfaces.ElectionTrigger
	BlockUtils          interfaces.BlockUtils
	KeyManager          *mocks.MockKeyManager
	Storage             interfaces.Storage
	Communication       *mocks.CommunicationMock
	Membership          interfaces.Membership
	MemberId            primitives.MemberId
	NodeStateChannel    chan *NodeState
	WriteToStateChannel bool
	PauseOnUpdateState  bool
	OnUpdateStateLatch  *test.Latch
}

func (node *Node) GetKeyManager() interfaces.KeyManager {
	return node.KeyManager
}

func (node *Node) GetMemberId() primitives.MemberId {
	return node.MemberId
}

func (node *Node) GetCurrentHeight() primitives.BlockHeight {
	return node.leanHelix.GetCurrentHeight()
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

func (node *Node) TriggerElectionOnNode(ctx context.Context) <-chan struct{} {

	electionTriggerMock, ok := node.ElectionTrigger.(*mocks.ElectionTriggerMock)
	if !ok {
		panic("You are trying to trigger election with an election trigger that is not the ElectionTriggerMock")
	}

	//node.leanHelix.TriggerElection(ctx, func(ctx context.Context) { electionTriggerMock.ManualTrigger(ctx) })
	fmt.Printf("Calling ManualTrigger on node %v\n", node.Membership.MyMemberId())
	return electionTriggerMock.ManualTrigger(ctx)
}

func (node *Node) onCommittedBlock(ctx context.Context, block interfaces.Block, blockProof []byte) {
	node.blockChain.AppendBlockToChain(block, blockProof)

	if node.WriteToStateChannel {
		nodeState := &NodeState{
			block: block,
		}

		select {
		case <-ctx.Done():
			return

		case node.NodeStateChannel <- nodeState:
			fmt.Printf("NODESTATE WROTE %v\n", block)
			return
		}
	}
}

func (node *Node) StartConsensus(ctx context.Context) {
	if node.leanHelix != nil {
		node.leanHelix.Run(ctx)
		node.leanHelix.UpdateState(ctx, node.GetLatestBlock(), nil)
	}
}

func (node *Node) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProof []byte, prevBlockProof []byte) error {
	return node.leanHelix.ValidateBlockConsensus(ctx, block, blockProof, prevBlockProof)
}

func (node *Node) Sync(ctx context.Context, prevBlock interfaces.Block, blockProofBytes []byte, prevBlockProofBytes []byte) {
	if node.leanHelix != nil {
		if err := node.ValidateBlockConsensus(ctx, prevBlock, blockProofBytes, prevBlockProofBytes); err == nil {
			go node.leanHelix.UpdateState(ctx, prevBlock, prevBlockProofBytes)
		}
	}
}

func (node *Node) SyncWithoutProof(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {
	node.leanHelix.UpdateState(ctx, prevBlock, prevBlockProofBytes)
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
	blockUtils interfaces.BlockUtils,
	electionTrigger interfaces.ElectionTrigger,
	logger interfaces.Logger) *Node {

	memberId := membership.MyMemberId()
	node := &Node{
		instanceId:          instanceId,
		blockChain:          mocks.NewInMemoryBlockChain(),
		ElectionTrigger:     electionTrigger,
		BlockUtils:          blockUtils,
		KeyManager:          mocks.NewMockKeyManager(memberId),
		Storage:             storage.NewInMemoryStorage(),
		Communication:       communication,
		Membership:          membership,
		MemberId:            memberId,
		NodeStateChannel:    make(chan *NodeState),
		OnUpdateStateLatch:  test.NewLatch(),
		WriteToStateChannel: true,
	}

	leanHelix := leanhelix.NewLeanHelix(node.BuildConfig(logger), node.onCommittedBlock)
	communication.RegisterOnMessage(leanHelix.HandleConsensusMessage)

	node.leanHelix = leanHelix
	return node

}
