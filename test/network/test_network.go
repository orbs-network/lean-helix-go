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
	"github.com/pkg/errors"
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

// Based on: https://stackoverflow.com/questions/52227954/waitgroup-on-subset-of-go-routines
type SubsetWaitGroup struct {
	remaining int
	mu        sync.RWMutex
	accChan   chan struct{}
}

func NewSubsetWaitGroup(total int, remaining int) *SubsetWaitGroup {
	if total-remaining <= 0 {
		panic("NewSubsetWaitGroup(): must have total > remaining")
	}
	return &SubsetWaitGroup{
		remaining: remaining,
		accChan:   make(chan struct{}, total-remaining),
	}
}

func (w *SubsetWaitGroup) Done() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.remaining--
	w.accChan <- struct{}{}
	return w.remaining
}

func (w *SubsetWaitGroup) Remaining() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.remaining
}

func (w *SubsetWaitGroup) WaitWithTimeout(timeout time.Duration) error {
	ch := make(chan struct{})
	timer := time.AfterFunc(timeout, func() {
		close(ch)
	})
	for doneCount := 0; doneCount < w.remaining; doneCount += 1 {
		select {
		case <-ch:
			return errors.Errorf("timed out with remaining=%d", w.Remaining())
		case <-w.accChan:
		}
	}
	timer.Stop()
	return nil
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

func (net *TestNetwork) WaitUntilNodesCommitASpecificBlock(ctx context.Context, t *testing.T, timeout time.Duration, block interfaces.Block, nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	if timeout == 0 {
		timeout = 2 * time.Second
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		var allEqual = true
		for _, node := range nodes {

			select {
			case <-timeoutCtx.Done():
				t.Fatal("WaitUntilNodesCommitASpecificBlock timed out")
				return
			case <-ctx.Done():
				return
			case nodeState := <-node.CommittedBlockChannel:
				if !matchers.BlocksAreEqual(block, nodeState.block) {
					allEqual = false
					//fmt.Printf("Expected: %s Committed: %s\n", block, nodeState.block)
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

	//fmt.Printf("---START---%d\n", len(nodes))
	wg := &sync.WaitGroup{}

	for _, node := range nodes {
		wg.Add(1)
		go func(node *Node) {
			if b := waitForAndReturnCommittedBlockAtHeight(ctx, node, block.Height()); b != nil { // NOTE - if ctx is cancelled we will never be wg.Done()
				if !matchers.BlocksAreEqual(block, b) {
					t.Fatalf("expected block at height %d to equal %v. found %v", block.Height(), block, b)
				}
				wg.Done()
			}
		}(node)
	}
	wg.Wait()
	//fmt.Printf("---DONE ALL---\n")

}

func waitForAndReturnCommittedBlockAtHeight(ctx context.Context, node *Node, targetHeight primitives.BlockHeight) interfaces.Block {
	nextItemToCheck := 0
	for ; ctx.Err() == nil; time.Sleep(10 * time.Millisecond) { // while context not cancelled
		var topBlockHeight primitives.BlockHeight
		if node.blockChain.LastBlock() != nil {
			topBlockHeight = node.blockChain.LastBlock().Height()
		}
		if topBlockHeight >= targetHeight { // if consensus reached for new blocks
			count := node.blockChain.Count()
			for ; nextItemToCheck < count; nextItemToCheck++ { // scan all newly appended blocks
				b, _ := node.blockChain.BlockAndProofAt(primitives.BlockHeight(nextItemToCheck))
				if b != nil && targetHeight == b.Height() { // if target height reached, return the block
					return b
				}
			}
		}
	}
	return nil
}

func doneAndSetZeroWhenReachingCount(wg *sync.WaitGroup, counter *int32) {
	//defer func() { _ = recover() }()
	wg.Done()
}

func (net *TestNetwork) WaitUntilQuorumCommitsHeight(ctx context.Context, height primitives.BlockHeight) {

	nodes := net.Nodes

	//fmt.Printf("---START---%d\n", len(nodes))
	//wg := &sync.WaitGroup{}

	// Should wait for "remaining" nodes and not for allNodeCount

	allNodeCount := len(nodes)
	var remaining int = quorum.CalcQuorumSize(allNodeCount)
	wg := NewSubsetWaitGroup(len(nodes), remaining)
	//wg.Add(allNodeCount)

	for _, node := range nodes {
		go func(node *Node) {
			for {
				var topBlock interfaces.Block
				select {
				case <-ctx.Done():
					wg.Done()
					//doneAndSetZeroWhenReachingCount(wg, &remaining)
					return
				case nodeState := <-node.CommittedBlockChannel:
					topBlock = nodeState.block
					//fmt.Printf("---READ--- ID=%s H=%d\n", node.MemberId, topBlock.Height())
				}
				//fmt.Printf("ID=%s Expected: %s Committed: %s\n", node.MemberId, height, topBlock.Height())
				if height == topBlock.Height() {
					//fmt.Printf("---DONE---%s\n", node.MemberId)
					//doneAndSetZeroWhenReachingCount(wg, &remaining)
					wg.Done()
					//fmt.Printf("---DONE---%s remaining=%d\n", node.MemberId, remaining)
					return
				}
			}
		}(node)
	}
	wg.WaitWithTimeout(1 * time.Second)
	//fmt.Printf("---DONE QUROM---\n")

}

func (net *TestNetwork) WaitUntilNodesCommitASpecificHeight(ctx context.Context, height primitives.BlockHeight, nodes ...*Node) {

	if nodes == nil {
		nodes = net.Nodes
	}

	//fmt.Printf("---START---%d\n", len(nodes))
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
					//fmt.Printf("---READ--- ID=%s H=%d\n", node.MemberId, topBlock.Height())
				}
				//fmt.Printf("ID=%s Expected: %s Committed: %s\n", node.MemberId, height, topBlock.Height())
				if height == topBlock.Height() {
					//fmt.Printf("---DONE---%s\n", node.MemberId)
					wg.Done()
					return
				}
			}
		}(node)
	}
	wg.Wait()
	//fmt.Printf("---DONE ALL---\n")

}

func (net *TestNetwork) WaitUntilNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight, nodes ...*Node) {

	if nodes == nil {
		nodes = net.Nodes
	}

	doneChan := make(chan struct{})
	for _, node := range nodes {
		go func(node *Node) {
			for {
				iterationTimeout, _ := context.WithTimeout(ctx, 20 * time.Millisecond)
				<- iterationTimeout.Done() // sleep for timeout

				if  ctx.Err() != nil { // parent context was cancelled, we're shutting down
					doneChan <- struct{}{}
					return
				}

				if node.GetCurrentHeight() >= height { // check height
					fmt.Printf("Node %s reached H=%d\n", node.MemberId, node.GetCurrentHeight())
					doneChan <- struct{}{}
					return
				}
			}
		}(node)
	}
	for i := 0; i < len(nodes); i++ {
		<- doneChan
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
