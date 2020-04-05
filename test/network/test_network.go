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
	log        interfaces.Logger
}

func (net *TestNetwork) GetNodeCommunication(memberId primitives.MemberId) *mocks.CommunicationMock {
	return net.Discovery.GetCommunicationById(memberId)
}

func (net *TestNetwork) StartConsensus(ctx context.Context) *TestNetwork {
	net.log.Debug("StartConsensus() start")
	for _, node := range net.Nodes {
		err := node.StartConsensus(ctx)
		if err != nil {
			panic(fmt.Sprintf("error starting consensus %s", err))
		}
	}

	net.WaitUntilNetworkIsRunning(ctx)

	net.log.Debug("StartConsensus: NETWORK IS READY")
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
	net.WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx, height, len(nodes), nodes...)

}

func (net *TestNetwork) WaitUntilQuorumOfNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight) {
	quorum := quorum.CalcQuorumSize(len(net.Nodes))
	net.WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx, height, quorum, net.Nodes...)
}
func (net *TestNetwork) WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx context.Context, height primitives.BlockHeight, subset int, nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}
	net.log.Debug("WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(): start: height=%d subset=%d numNodes=%d ", height, subset, len(nodes))
	doneChan := make(chan struct{})
	for _, node := range nodes {
		// TODO Trying to use node.blockChain.Count() instead of node.GetCurrentHeight()
		// hangs - need to understand why, as it seems more correct.

		go func(node *Node) { // shadowing node on purpose
			for node.GetCurrentHeight() < height {
				iterationTimeout, _ := context.WithTimeout(ctx, 20*time.Millisecond)
				<-iterationTimeout.Done() // sleep or get cancelled
				if ctx.Err() != nil {
					return
				}
			}

			select {
			case doneChan <- struct{}{}:
			case <-ctx.Done():
			}
		}(node)
	}
	for i := 0; i < subset; i++ {
		select {
		case <-doneChan:
		case <-ctx.Done():
		}
	}
	net.log.Debug("WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(): end: height=%d subset=%d numNodes=%d ", height, subset, len(nodes))
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
		node.SetRequestNewBlockCallsLeftUntilItPausesWhenCounterIsZero(counter)
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

func (net *TestNetwork) TriggerElectionsOnNodes(ctx context.Context, nodes ...*Node) {
	for _, n := range nodes {
		<-n.TriggerElectionOnNode(ctx)
	}
}

func (net *TestNetwork) TriggerElectionsOnAllNodes(ctx context.Context) {
	net.TriggerElectionsOnNodes(ctx, net.Nodes...)
}

func (net *TestNetwork) WaitForShutdown(ctx context.Context) {
	net.log.Debug("WaitForShutdown() start")
	for _, node := range net.Nodes {
		node.leanHelix.WaitUntilShutdown(ctx)
	}
	net.log.Debug("WaitForShutdown() DONE")
}

func NewTestNetwork(instanceId primitives.InstanceId, discovery *mocks.Discovery, log interfaces.Logger) *TestNetwork {
	return &TestNetwork{
		InstanceId: instanceId,
		Nodes:      []*Node{},
		Discovery:  discovery,
		log:        log,
	}
}
