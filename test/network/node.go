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
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"time"
)

type NodeState struct {
	block           interfaces.Block
	blockProofBytes []byte
	validationCount int
}

type Node struct {
	instanceId               primitives.InstanceId
	leanHelix                *leanhelix.MainLoop
	blockChain               *mocks.InMemoryBlockChain
	ElectionTrigger          interfaces.ElectionScheduler
	BlockUtils               interfaces.BlockUtils
	KeyManager               *mocks.MockKeyManager
	Storage                  interfaces.Storage
	Communication            *mocks.CommunicationMock
	Membership               interfaces.Membership
	MemberId                 primitives.MemberId
	CommittedBlockChannel    chan *NodeState
	NewConsensusRoundChannel chan primitives.BlockHeight
	WriteToStateChannel      bool
	OnUpdateStateLatch       *test.Latch
}

func (node *Node) State() state.State {
	return node.leanHelix.State()
}

func (node *Node) GetKeyManager() interfaces.KeyManager {
	return node.KeyManager
}

func (node *Node) GetMemberId() primitives.MemberId {
	return node.MemberId
}

func (node *Node) GetCurrentHeight() primitives.BlockHeight {
	return node.leanHelix.State().Height()
}

func (node *Node) GetLatestBlock() interfaces.Block {
	return node.blockChain.LastBlock()
}

func (node *Node) GetLatestBlockProof() []byte {
	return node.blockChain.LastBlockProof()
}

func (node *Node) GetBlockProofAt(height primitives.BlockHeight) []byte {
	return node.blockChain.BlockProofAt(height)
}

func (node *Node) TriggerElectionOnNode(ctx context.Context) <-chan struct{} {

	electionTriggerMock, ok := node.ElectionTrigger.(*mocks.ElectionTriggerMock)
	if !ok {
		panic("You are trying to trigger election with an election trigger that is not the ElectionTriggerMock")
	}

	return electionTriggerMock.ManualTrigger(ctx, node.State().HeightView())
}

func (node *Node) onCommittedBlock(ctx context.Context, block interfaces.Block, blockProof []byte) {
	node.blockChain.AppendBlockToChain(block, blockProof)

	if node.WriteToStateChannel {
		nodeState := &NodeState{
			block:           block,
			blockProofBytes: blockProof,
		}

		select {
		case <-ctx.Done():
			return

		case node.CommittedBlockChannel <- nodeState:
			return
		}
	}
}

func (node *Node) BlockChain() *mocks.InMemoryBlockChain {
	return node.blockChain
}

func (node *Node) onNewConsensusRound(ctx context.Context, newHeight primitives.BlockHeight, prevBlock interfaces.Block, canBeFirstLeader bool) {

	if node.NewConsensusRoundChannel == nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	case node.NewConsensusRoundChannel <- newHeight:
		return
	}
}

func (node *Node) PauseOnNewConsensusRoundUntilReadingFrom(c chan primitives.BlockHeight) *Node {
	node.NewConsensusRoundChannel = c
	return node
}

func (node *Node) DontPauseOnNewConsensusRound() *Node {
	node.NewConsensusRoundChannel = nil
	return node
}

func (node *Node) ConsensusRoundChannel() chan primitives.BlockHeight {
	return node.NewConsensusRoundChannel
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
	if node.leanHelix == nil {
		return
	}
	fmt.Printf("ID=%s H=%d B.H=%d NodeSync(): Start, calling ValidateBlockConsensus\n", node.MemberId, node.GetCurrentHeight(), prevBlock.Height())
	if err := node.ValidateBlockConsensus(ctx, prevBlock, blockProofBytes, prevBlockProofBytes); err == nil {
		fmt.Printf("ID=%s H=%d B.H=%d NodeSync(): Passed ValidateBlockConsensus, calling UpdateState\n", node.MemberId, node.GetCurrentHeight(), prevBlock.Height())
		node.leanHelix.UpdateState(ctx, prevBlock, prevBlockProofBytes)
		fmt.Printf("ID=%s H=%d B.H=%d NodeSync(): UpdateState returned\n", node.MemberId, node.GetCurrentHeight(), prevBlock.Height())
	} else {
		fmt.Printf("ID=%s H=%d B.H=%d NodeSync(): Failed validation: %s\n", node.MemberId, node.GetCurrentHeight(), prevBlock.Height(), err)
	}
}

func (node *Node) SyncWithoutValidation(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {
	node.leanHelix.UpdateState(ctx, prevBlock, prevBlockProofBytes)
}

func (node *Node) BuildConfig(logger interfaces.Logger) *interfaces.Config {
	return &interfaces.Config{
		InstanceId:            node.instanceId,
		Communication:         node.Communication,
		Membership:            node.Membership,
		BlockUtils:            node.BlockUtils,
		KeyManager:            node.KeyManager,
		ElectionTimeoutOnV0:   10 * time.Millisecond,
		OnElectionCB:          nil,
		Storage:               node.Storage,
		Logger:                logger,
		MsgChanBufLen:         10,
		UpdateStateChanBufLen: 10,
		ElectionChanBufLen:    0,
	}

}

func NewNode(
	instanceId primitives.InstanceId,
	membership interfaces.Membership,
	communication *mocks.CommunicationMock,
	blockUtils interfaces.BlockUtils,
	electionTrigger interfaces.ElectionScheduler,
	logger interfaces.Logger) *Node {

	if electionTrigger == nil {
		electionTrigger = mocks.NewMockElectionTrigger()
	}
	memberId := membership.MyMemberId()

	node := &Node{
		instanceId:               instanceId,
		blockChain:               mocks.NewInMemoryBlockChain(),
		ElectionTrigger:          electionTrigger,
		BlockUtils:               blockUtils,
		KeyManager:               mocks.NewMockKeyManager(memberId),
		Storage:                  storage.NewInMemoryStorage(),
		Communication:            communication,
		Membership:               membership,
		MemberId:                 memberId,
		CommittedBlockChannel:    make(chan *NodeState),
		NewConsensusRoundChannel: nil,
		OnUpdateStateLatch:       test.NewLatch(),
		WriteToStateChannel:      true,
	}
	config := node.BuildConfig(logger)
	config.OverrideElectionTrigger = node.ElectionTrigger

	leanHelix := leanhelix.NewLeanHelix(config, node.onCommittedBlock, node.onNewConsensusRound)
	communication.RegisterIncomingMessageHandler(leanHelix.HandleConsensusMessage)

	node.leanHelix = leanHelix
	return node

}
