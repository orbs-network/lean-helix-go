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

func (net *TestNetwork) WaitUntilNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight, nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}
	net._waitUntilQuorumOfNodesEventuallyReachASpecificHeight(ctx, height, len(nodes), nodes...)
}

func (net *TestNetwork) WaitUntilQuorumOfNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight) {
	quorum := quorum.CalcQuorumSize(len(net.Nodes))
	net._waitUntilQuorumOfNodesEventuallyReachASpecificHeight(ctx, height, quorum, net.Nodes...)
}
func (net *TestNetwork) _waitUntilQuorumOfNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight, quorum int, nodes ...*Node) {
	doneChan := make(chan struct{})
	for _, node := range nodes {
		// TODO Trying to use node.blockChain.Count() instead of node.GetCurrentHeight()
		// hangs - need to understand why, as it seems more correct.

		go func(node *Node) { // shadowing node on purpose
			defer func() {
				if e := recover(); e != nil {
					s, ok := e.(error)
					if ok && s.Error() != "send on closed channel" {
						node.log.Debug("H=%d ID=%s _waitUntilQuorumOfNodesEventuallyReachASpecificHeight exited with error: %s", node.GetCurrentHeight(), node.Membership.MyMemberId(), height, s)
					}
					if ok && s.Error() == "send on closed channel" {
						node.log.Debug("H=%d ID=%s _waitUntilQuorumOfNodesEventuallyReachASpecificHeight exited with error: %s AFTER QUORUM REACHED", node.GetCurrentHeight(), node.Membership.MyMemberId(), height, s)
					}
					if !ok {
						node.log.Debug("H=%d ID=%s _waitUntilQuorumOfNodesEventuallyReachASpecificHeight exited with error: %v UNKNOWN ERROR", node.GetCurrentHeight(), node.Membership.MyMemberId(), height, e)
					}
				}
			}()
			for node.GetCurrentHeight() < height {
				iterationTimeout, _ := context.WithTimeout(ctx, 20*time.Millisecond)
				<-iterationTimeout.Done() // sleep or get cancelled

				if ctx.Err() != nil {
					break // shutting down
				}
			}
			if height > 1 {
				node.log.Debug("H=%d ID=%s _waitUntilQuorumOfNodesEventuallyReachASpecificHeight (target height %d)", node.GetCurrentHeight(), node.Membership.MyMemberId(), height)
			}

			doneChan <- struct{}{}
		}(node)
	}
	for i := 0; i < quorum; i++ {
		<-doneChan
	}
	close(doneChan)
}

// Wait for H=1 so that election triggers will be sent with H=1
// o/w they will sometimes be sent with H=0 and subsequently be ignored
// by workerloop's election channel, causing election to not happen,
// failing/hanging the test.
func (net *TestNetwork) WaitUntilNetworkIsRunning(ctx context.Context) {
	net.WaitUntilNodesEventuallyReachASpecificHeight(ctx, 1)
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
