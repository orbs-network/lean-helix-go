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
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"math"
	"sync"
	"testing"
	"time"
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
		err := node.StartConsensus(ctx)
		if err != nil {
			panic(fmt.Sprintf("error starting consensus %s", err))
		}
	}

	for _, node := range net.Nodes {
		net.WaitUntilCurrentHeightGreaterEqualThan(ctx, 1, node)
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

const MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS = 4

func (net *TestNetwork) MAYBE_FLAKY_WaitForAllNodesToCommitTheSameBlock(ctx context.Context) bool {
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

// TODO this function hides the fact that nodes don't necessarily produced the same block. and we old blocks may also be returned. the last node's block is always returned and all the other's ignored.
func (net *TestNetwork) WaitUntilNodesCommitAnyBlock(ctx context.Context, nodes ...*Node) interfaces.Block {
	if nodes == nil {
		nodes = net.Nodes
	}

	var nodeState *NodeState = nil

	for _, node := range nodes {
		select {
		case <-ctx.Done():
			return nil
		case nodeState = <-node.CommittedBlockChannel:
			continue
		}
	}

	if nodeState == nil {
		return nil
	}
	return nodeState.block
}

func (net *TestNetwork) WaitUntilNodesCommitASpecificBlock(ctx context.Context, block interfaces.Block, nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for {
		var allEqual = true
		for _, node := range nodes {

			select {
			case <-ctx.Done():
				return
			case nodeState := <-node.CommittedBlockChannel:
				if !matchers.BlocksAreEqual(block, nodeState.block) {
					allEqual = false
					fmt.Printf("Expected: %s Committed: %s\n", block, nodeState.block)
					break
				}
			}
		}
		if allEqual {
			return
		}
	}
}

func (net *TestNetwork) WaitUntilNodesEventuallyCommitASpecificBlock(ctx context.Context, t *testing.T, block interfaces.Block, nodes ...*Node) {

	if nodes == nil {
		nodes = net.Nodes
	}

	fmt.Printf("---START---%d\n", len(nodes))
	wg := &sync.WaitGroup{}

	for _, node := range nodes {
		wg.Add(1)
		go func(node *Node) {
			if b := waitForCommittedBlockAtHeight(ctx, node, block.Height()); b != nil { // NOTE - if ctx is cancelled we will never be wg.Done()
				if !matchers.BlocksAreEqual(block, b) {
					t.Fatalf("expected block at height %d to equal %v. found %v", block.Height(), block, b)
				}
				wg.Done()
			}
		}(node)
	}
	wg.Wait()
	fmt.Printf("---DONE ALL---\n")

}

func waitForCommittedBlockAtHeight(ctx context.Context, node *Node, targetHeight primitives.BlockHeight) interfaces.Block {
	nextItemToCheck := 0
	for ; ctx.Err() == nil; time.Sleep(10 * time.Millisecond) { // while context not cancelled
		var topBlockHeight primitives.BlockHeight
		if node.blockChain.LastBlock() != nil {
			topBlockHeight = node.blockChain.LastBlock().Height()
		}
		if topBlockHeight >= targetHeight { // if consensus reached for new blocks
			for ; nextItemToCheck < len(node.blockChain.Items()); nextItemToCheck++ { // scan all newly appended blocks
				b := node.blockChain.Items()[nextItemToCheck]
				if b.Block() != nil && targetHeight == b.Block().Height() { // if target height reached, return the block
					return b.Block()
				}
			}
		}
	}
	return nil
}

func (net *TestNetwork) WaitUntilQuorumCommitsHeight(ctx context.Context, height primitives.BlockHeight) {

	nodes := net.Nodes

	fmt.Printf("---START---%d\n", len(nodes))
	wg := &sync.WaitGroup{}

	wg.Add(quorum.CalcQuorumSize(len(nodes)))

	for _, node := range nodes {
		go func(node *Node) {
			for {
				var topBlock interfaces.Block
				select {
				case <-ctx.Done():
					return
				case nodeState := <-node.CommittedBlockChannel:
					topBlock = nodeState.block
					fmt.Printf("---READ--- ID=%s H=%d\n", node.MemberId, topBlock.Height())
				}
				fmt.Printf("ID=%s Expected: %s Committed: %s\n", node.MemberId, height, topBlock.Height())
				if height == topBlock.Height() {
					fmt.Printf("---DONE---%s\n", node.MemberId)
					go func() {
						defer func() { _ = recover() }()
						wg.Done()
					}() // may panic but we're cool
					return
				}
			}
		}(node)
	}
	wg.Wait()
	fmt.Printf("---DONE QUROM---\n")

}

func (net *TestNetwork) WaitUntilNodesCommitASpecificHeight(ctx context.Context, height primitives.BlockHeight, nodes ...*Node) {

	if nodes == nil {
		nodes = net.Nodes
	}

	fmt.Printf("---START---%d\n", len(nodes))
	wg := &sync.WaitGroup{}

	for _, node := range nodes {
		wg.Add(1)
		go func(node *Node) {
			for {
				var topBlock interfaces.Block
				select {
				case <-ctx.Done():
					wg.Done()
					return
				case nodeState := <-node.CommittedBlockChannel:
					topBlock = nodeState.block
					fmt.Printf("---READ--- ID=%s H=%d\n", node.MemberId, topBlock.Height())
				}
				fmt.Printf("ID=%s Expected: %s Committed: %s\n", node.MemberId, height, topBlock.Height())
				if height == topBlock.Height() {
					fmt.Printf("---DONE---%s\n", node.MemberId)
					wg.Done()
					return
				}
			}
		}(node)
	}
	wg.Wait()
	fmt.Printf("---DONE ALL---\n")

}

func (net *TestNetwork) WaitUntilNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight, nodes ...*Node) {

	if nodes == nil {
		nodes = net.Nodes
	}

	fmt.Printf("---START---%d\n", len(nodes))
	wg := &sync.WaitGroup{}

	for _, node := range nodes {
		wg.Add(1)
		go func(node *Node) {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				default:
					if node.GetCurrentHeight() >= height {
						wg.Done()
						return
					}
					time.Sleep(20 * time.Millisecond)
				}
			}
		}(node)
	}
	wg.Wait()
	fmt.Printf("---DONE ALL---\n")

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
			pausableBlockUtils.RequestNewBlockCallsLeftUntilItPausesWhenCounterIsZero = counter
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

func (net *TestNetwork) WaitUntilCurrentHeightGreaterEqualThan(ctx context.Context, height primitives.BlockHeight, node *Node) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if node.GetCurrentHeight() >= height {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func (net *TestNetwork) WaitUntilNewConsensusRoundForBlockHeight(ctx context.Context, height primitives.BlockHeight, node *Node) {
	if node.GetCurrentHeight() >= height {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case h := <-node.ConsensusRoundChannel():
			fmt.Printf("")
			if h == height {
				return
			}
		}
	}
}

func NewTestNetwork(instanceId primitives.InstanceId, discovery *mocks.Discovery) *TestNetwork {
	return &TestNetwork{
		InstanceId: instanceId,
		Nodes:      []*Node{},
		Discovery:  discovery,
	}
}
