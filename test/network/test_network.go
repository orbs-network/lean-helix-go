// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package network

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"math"
)

type TestNetwork struct {
	InstanceId primitives.InstanceId
	Nodes      []*Node
	Discovery  *mocks.Discovery
}

func (net *TestNetwork) GetNodeCommunication(memberId primitives.MemberId) *mocks.CommunicationMock {
	return net.Discovery.GetCommunicationById(memberId)
}

func (net *TestNetwork) StartConsensus(ctx context.Context) *TestNetwork {
	for _, node := range net.Nodes {
		node.StartConsensus(ctx)
	}

	return net
}

func (net *TestNetwork) RegisterNode(node *Node) {
	net.Nodes = append(net.Nodes, node)
}

func (net *TestNetwork) RegisterNodes(nodes []*Node) {
	for _, node := range nodes {
		net.RegisterNode(node)
	}
}

func (net *TestNetwork) AllNodesChainEndsWithABlock(block interfaces.Block) bool {
	for _, node := range net.Nodes {
		if matchers.BlocksAreEqual(block, node.GetLatestBlock()) == false {
			return false
		}
	}
	return true
}

func (net *TestNetwork) WaitForAllNodesToCommitBlockAndReturnWhetherEqualToGiven(ctx context.Context, expectedBlock interfaces.Block) bool {
	for _, node := range net.Nodes {
		select {
		case <-ctx.Done():
			return false
		case nodeState := <-node.CommittedBlockChannel:
			blockAreEqual := matchers.BlocksAreEqual(expectedBlock, nodeState.block)
			if blockAreEqual == false {
				return false
			}
		}
	}
	return true
}

const MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS = 4

func (net *TestNetwork) WaitForAllNodesToCommitTheSameBlock(ctx context.Context) bool {
	if len(net.Nodes) < MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS {
		panic("Not enough nodes for consensus")
	}

	select {
	case <-ctx.Done():
		return false
	case firstNodeStateChannel := <-net.Nodes[0].CommittedBlockChannel:
		firstNodeBlock := firstNodeStateChannel.block
		for i := 1; i < len(net.Nodes); i++ {
			node := net.Nodes[i]

			select {
			case <-ctx.Done():
				return false

			case nodeState := <-node.CommittedBlockChannel:
				if matchers.BlocksAreEqual(firstNodeBlock, nodeState.block) == false {
					return false
				}
			}

		}
		return true
	}
}

func (net *TestNetwork) SetNodesToPauseOnValidateBlock(nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		if pausableBlockUtils, ok := node.BlockUtils.(*mocks.PausableBlockUtils); ok {
			pausableBlockUtils.PauseOnValidateBlock = true
		} else {
			panic("Node.BlockUtils is not PausableBlockUtils")
		}

	}
}

func (net *TestNetwork) ReturnWhenNodesPauseOnValidateBlock(ctx context.Context, nodes ...*Node) {
	for _, node := range nodes {
		if pausableBlockUtils, ok := node.BlockUtils.(*mocks.PausableBlockUtils); ok {
			pausableBlockUtils.ValidationLatch.ReturnWhenLatchIsPaused(ctx, node.MemberId)
		} else {
			panic("Node.BlockUtils is not PausableBlockUtils")
		}

	}
}

func (net *TestNetwork) ResumeValidateBlockOnNodes(ctx context.Context, nodes ...*Node) {
	for _, node := range nodes {
		if pausableBlockUtils, ok := node.BlockUtils.(*mocks.PausableBlockUtils); ok {
			pausableBlockUtils.ValidationLatch.Resume(ctx, node.MemberId)
		} else {
			panic("Node.BlockUtils is not PausableBlockUtils")
		}

	}
}

func (net *TestNetwork) SetNodesToPauseOnRequestNewBlock(nodes ...*Node) {
	net.SetNodesPauseOnRequestNewBlockWhenCounterIsZero(0, nodes...)
}

func (net *TestNetwork) SetNodesToNotPauseOnRequestNewBlock(nodes ...*Node) {
	net.SetNodesPauseOnRequestNewBlockWhenCounterIsZero(math.MaxInt64, nodes...)
}

func (net *TestNetwork) SetNodesPauseOnRequestNewBlockWhenCounterIsZero(counter int64, nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}
	for _, node := range nodes {
		if pausableBlockUtils, ok := node.BlockUtils.(*mocks.PausableBlockUtils); ok {
			pausableBlockUtils.PauseOnRequestNewBlockWhenCounterIsZero = counter
		} else {
			panic("Node.BlockUtils is not PausableBlockUtils")
		}
	}
}

func (net *TestNetwork) ReturnWhenNodeIsPausedOnRequestNewBlock(ctx context.Context, node *Node) {
	if pausableBlockUtils, ok := node.BlockUtils.(*mocks.PausableBlockUtils); ok {
		pausableBlockUtils.RequestNewBlockLatch.ReturnWhenLatchIsPaused(ctx, node.MemberId)
	} else {
		panic("Node.BlockUtils is not PausableBlockUtils")
	}
}

func (net *TestNetwork) ReturnWhenNodesPauseOnUpdateState(ctx context.Context, node *Node) {
	node.OnUpdateStateLatch.ReturnWhenLatchIsPaused(ctx, node.MemberId)
}

func (net *TestNetwork) ResumeRequestNewBlockOnNodes(ctx context.Context, node *Node) {
	if pausableBlockUtils, ok := node.BlockUtils.(*mocks.PausableBlockUtils); ok {
		pausableBlockUtils.RequestNewBlockLatch.Resume(ctx, node.MemberId)
	} else {
		panic("Node.BlockUtils is not PausableBlockUtils")
	}

}

func (net *TestNetwork) WaitForNodesToCommitABlock(ctx context.Context, nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		fmt.Printf("ID=%s WaitForNodesToCommitABlock BEFORE\n", node.MemberId)
		select {
		case <-ctx.Done():
			fmt.Printf("ID=%s WaitForNodesToCommitABlock ctx.Done\n", node.MemberId)
			return
		case <-node.CommittedBlockChannel:
			fmt.Printf("ID=%s WaitForNodesToCommitABlock AFTER\n", node.MemberId)
			continue
		}
	}
}

func (net *TestNetwork) WaitForNodesToCommitASpecificBlock(ctx context.Context, block interfaces.Block, nodes ...*Node) bool {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {

		select {
		case <-ctx.Done():
			return false
		case nodeState := <-node.CommittedBlockChannel:
			if matchers.BlocksAreEqual(block, nodeState.block) == false {
				return false
			}
		}

	}
	return true
}

func (net *TestNetwork) AllNodesValidatedNoMoreThanOnceBeforeCommit(ctx context.Context) bool {
	for _, node := range net.Nodes {

		select {
		case <-ctx.Done():
			return false
		case nodeState := <-node.CommittedBlockChannel:
			if nodeState.validationCount > 1 {
				return false
			}
		}
	}
	return true
}

func (net *TestNetwork) SetNodesToNotPauseOnTheFirstXTimesOfOnRequestNewBlock(timesNotToPause int) {

}

func NewTestNetwork(instanceId primitives.InstanceId, discovery *mocks.Discovery) *TestNetwork {
	return &TestNetwork{
		InstanceId: instanceId,
		Nodes:      []*Node{},
		Discovery:  discovery,
	}
}
