// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package network

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/electiontrigger"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"testing"
	"time"
)

type TestNetworkBuilder struct {
	instanceId                          primitives.InstanceId
	NodeCount                           int
	logger                              interfaces.Logger
	upcomingBlocks                      []interfaces.Block
	keyManager                          interfaces.KeyManager
	blockUtils                          []interfaces.BlockUtils
	communication                       interfaces.Communication
	orderCommitteeByHeight              bool
	communicationMaxDelay               time.Duration
	electionTriggerTimeout              time.Duration
	useTimeBasedElectionTrigger         bool
	withFailingBlockProposalValidations bool
}

func (tb *TestNetworkBuilder) WithNodeCount(nodeCount int) *TestNetworkBuilder {
	tb.NodeCount = nodeCount
	return tb
}

func (tb *TestNetworkBuilder) InNetwork(instanceId primitives.InstanceId) *TestNetworkBuilder {
	tb.instanceId = instanceId
	return tb
}

func (tb *TestNetworkBuilder) WithBlocks(upcomingBlocks ...interfaces.Block) *TestNetworkBuilder {
	if tb.upcomingBlocks == nil {
		tb.upcomingBlocks = upcomingBlocks
	}
	return tb
}

func (tb *TestNetworkBuilder) WithTimeBasedElectionTrigger(timeout time.Duration) *TestNetworkBuilder {
	tb.useTimeBasedElectionTrigger = true
	tb.electionTriggerTimeout = timeout
	return tb
}

func (tb *TestNetworkBuilder) GossipMessagesMaxDelay(duration time.Duration) *TestNetworkBuilder {
	tb.communicationMaxDelay = duration
	return tb
}

func (tb *TestNetworkBuilder) LogToConsole(t testing.TB) *TestNetworkBuilder {
	tb.logger = logger.NewConsoleLogger(test.NameHashPrefix(t, 4))
	return tb
}

func (tb *TestNetworkBuilder) OrderCommitteeByHeight() *TestNetworkBuilder {
	tb.orderCommitteeByHeight = true
	return tb
}

func (tb *TestNetworkBuilder) Build(ctx context.Context) *TestNetwork {

	if tb.logger == nil {
		tb.logger = logger.NewSilentLogger()
	}

	blocksPool := tb.buildBlocksPool()
	discovery := mocks.NewDiscovery()
	nodes := tb.createNodes(discovery, blocksPool)
	testNetwork := NewTestNetwork(tb.instanceId, discovery)
	testNetwork.RegisterNodes(nodes)

	tb.setupCommChannels(ctx, testNetwork)

	return testNetwork
}

func (tb *TestNetworkBuilder) setupCommChannels(ctx context.Context, network *TestNetwork) {
	for _, node := range network.Nodes {
		for _, peerNode := range network.Nodes {
			comm := network.GetNodeCommunication(node.MemberId)
			comm.ReturnAndMaybeCreateOutgoingChannelByTarget(ctx, peerNode.MemberId)
		}
	}
}

func (tb *TestNetworkBuilder) buildBlocksPool() *mocks.BlocksPool {
	if tb.upcomingBlocks == nil {
		b1 := mocks.ABlock(interfaces.GenesisBlock)
		b2 := mocks.ABlock(b1)
		b3 := mocks.ABlock(b2)
		b4 := mocks.ABlock(b3)

		return mocks.NewBlocksPool([]interfaces.Block{b1, b2, b3, b4})
	} else {
		return mocks.NewBlocksPool(tb.upcomingBlocks)
	}
}

func (tb *TestNetworkBuilder) buildNode(
	nodeBuilder *NodeBuilder,
	memberId primitives.MemberId,
	discovery *mocks.Discovery,
	blockUtils interfaces.BlockUtils,
) *Node {

	communicationInstance := mocks.NewCommunication(memberId, discovery)
	if tb.communicationMaxDelay > 0 {
		communicationInstance.SetMessagesMaxDelay(tb.communicationMaxDelay)
	}
	discovery.RegisterCommunication(memberId, communicationInstance)
	membership := mocks.NewFakeMembership(memberId, discovery, tb.orderCommitteeByHeight)

	b := nodeBuilder.
		AsInstanceId(tb.instanceId).
		CommunicatesVia(communicationInstance).
		ThatIsPartOf(membership).
		WithBlockUtils(blockUtils).
		WithMemberId(memberId).
		WithLogger(tb.logger)

	if tb.useTimeBasedElectionTrigger {
		et := Electiontrigger.NewTimerBasedElectionTrigger(tb.electionTriggerTimeout, nil)
		b.WithElectionTrigger(et)
	}
	return b.Build()
}

func (tb *TestNetworkBuilder) createNodes(discovery *mocks.Discovery, blocksPool *mocks.BlocksPool) []*Node {
	var nodes []*Node
	for i := 0; i < tb.NodeCount; i++ {

		memberId := primitives.MemberId(fmt.Sprintf("%03d", i))
		var blockUtils interfaces.BlockUtils
		if i < len(tb.blockUtils) {
			blockUtils = tb.blockUtils[i]
		} else {
			pausableBlockUtils := mocks.NewMockBlockUtils(memberId, blocksPool, tb.logger)
			if tb.withFailingBlockProposalValidations {
				pausableBlockUtils.WithFailingBlockProposalValidations()
			}
			blockUtils = pausableBlockUtils
		}

		nodeBuilder := NewNodeBuilder()
		node := tb.buildNode(nodeBuilder, memberId, discovery, blockUtils)
		nodes = append(nodes, node)
	}

	return nodes
}

func (tb *TestNetworkBuilder) WithCommunication(communication interfaces.Communication) *TestNetworkBuilder {
	tb.communication = communication
	return tb
}

func (tb *TestNetworkBuilder) WithBlockUtils(utils []interfaces.BlockUtils) *TestNetworkBuilder {
	tb.blockUtils = utils
	return tb
}

func (tb *TestNetworkBuilder) WithKeyManager(mgr interfaces.KeyManager) *TestNetworkBuilder {
	tb.keyManager = mgr
	return tb
}

func (tb *TestNetworkBuilder) WithMaybeFailingBlockProposalValidations(b bool, blocksPool ...interfaces.Block) *TestNetworkBuilder {
	tb.withFailingBlockProposalValidations = b
	tb.upcomingBlocks = blocksPool
	return tb
}

func NewTestNetworkBuilder() *TestNetworkBuilder {
	return &TestNetworkBuilder{
		NodeCount:      0,
		upcomingBlocks: nil,
	}
}

func ABasicTestNetwork(ctx context.Context) *TestNetwork {
	return ATestNetworkBuilder(4).
		Build(ctx)
}

func ABasicTestNetworkWithConsoleLogs(ctx context.Context, tb testing.TB) *TestNetwork {
	return ATestNetworkBuilder(4).
		LogToConsole(tb).
		Build(ctx)
}

func ATestNetworkBuilder(countOfNodes int, blocksPool ...interfaces.Block) *TestNetworkBuilder {
	//instanceId := primitives.InstanceId(rand.Uint64())
	return NewTestNetworkBuilder().
		WithNodeCount(countOfNodes).
		WithBlocks(blocksPool...)
	//InNetwork(instanceId) // generates a random InstanceID, not needed for most tests

}
