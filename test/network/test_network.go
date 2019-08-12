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

	remaining := w.DecrementRemaining()
	w.accChan <- struct{}{}
	return remaining
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
	remaining := w.Remaining()
	for doneCount := 0; doneCount < remaining; doneCount += 1 {
		select {
		case <-ch:
			return errors.Errorf("timed out with remaining=%d", remaining)
		case <-w.accChan:
		}
	}
	timer.Stop()
	return nil
}

func (w *SubsetWaitGroup) DecrementRemaining() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.remaining--
	return w.remaining
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

	net.WaitUntilNetworkIsRunning(ctx)
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

func (net *TestNetwork) WaitUntilNodesEventuallyCommitASpecificBlock(ctx context.Context, t *testing.T, timeout time.Duration, block interfaces.Block, nodes ...*Node) {

	if nodes == nil {
		nodes = net.Nodes
	}

	h := block.Height()
	net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, h+1, nodes...)
	for _, node := range nodes {
		b, _ := node.blockChain.BlockAndProofAt(h)
		if !matchers.BlocksAreEqual(block, b) {
			t.Fatalf("Node %s: Height=%d: Expected block %s but found %s",
				node.MemberId, h, block, b)
		}
	}
}

func (net *TestNetwork) WaitUntilQuorumCommitsHeight(ctx context.Context, height primitives.BlockHeight) {

	nodes := net.Nodes

	// Should wait for "remaining" nodes and not for allNodeCount
	allNodeCount := len(nodes)
	var remaining int = quorum.CalcQuorumSize(allNodeCount)
	wg := NewSubsetWaitGroup(len(nodes), remaining)

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
				}
				if height == topBlock.Height() {
					wg.Done()
					return
				}
			}
		}(node)
	}
	wg.WaitWithTimeout(1 * time.Second)
}

// Wait for H=1 so that election triggers will be sent with H=1
// o/w they will sometimes be sent with H=0 and subsequently be ignored
// by workerloop's election channel, causing election to not happen,
// failing/hanging the test.
func (net *TestNetwork) WaitUntilNetworkIsRunning(ctx context.Context) {
	net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 1)
}

func (net *TestNetwork) WaitUntilNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight, nodes ...*Node) {

	if nodes == nil {
		nodes = net.Nodes
	}

	doneChan := make(chan struct{})
	for _, node := range nodes {
		// TODO Trying to use node.blockChain.Count() instead of node.GetCurrentHeight()
		// hangs - need to understand why, as it seems more correct.

		go func(n *Node) {
			for n.GetCurrentHeight() < height {
				iterationTimeout, _ := context.WithTimeout(ctx, 20*time.Millisecond)
				<-iterationTimeout.Done() // sleep or get cancelled

				if ctx.Err() != nil {
					break // shutting down
				}
			}
			fmt.Printf("Node %s reached H=%d\n", n.MemberId, n.GetCurrentHeight())
			doneChan <- struct{}{}
		}(node)
	}
	for i := 0; i < len(nodes); i++ {
		<-doneChan
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

func (net *TestNetwork) TriggerElectionsOnAllNodes(ctx context.Context) {
	for _, n := range net.Nodes {
		<-n.TriggerElectionOnNode(ctx)
	}
}

func NewTestNetwork(instanceId primitives.InstanceId, discovery *mocks.Discovery) *TestNetwork {
	return &TestNetwork{
		InstanceId: instanceId,
		Nodes:      []*Node{},
		Discovery:  discovery,
	}
}
