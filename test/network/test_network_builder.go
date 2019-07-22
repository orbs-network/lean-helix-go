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
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"math/rand"
	"time"
)

type TestNetworkBuilder struct {
	instanceId                          primitives.InstanceId
	NodeCount                           int
	logToConsole                        bool
	customNodeBuilders                  []*NodeBuilder
	upcomingBlocks                      []interfaces.Block
	keyManager                          interfaces.KeyManager
	blockUtils                          interfaces.BlockUtils
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

func (tb *TestNetworkBuilder) WithCustomNodeBuilder(nodeBuilder *NodeBuilder) *TestNetworkBuilder {
	tb.customNodeBuilders = append(tb.customNodeBuilders, nodeBuilder)
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

func (tb *TestNetworkBuilder) LogToConsole() *TestNetworkBuilder {
	tb.logToConsole = true
	return tb
}

func (tb *TestNetworkBuilder) OrderCommitteeByHeight() *TestNetworkBuilder {
	tb.orderCommitteeByHeight = true
	return tb
}

func (tb *TestNetworkBuilder) Build(ctx context.Context) *TestNetwork {
	blocksPool := tb.buildBlocksPool()
	discovery := mocks.NewDiscovery()
	nodes := tb.createNodes(discovery, blocksPool, tb.logToConsole, tb.withFailingBlockProposalValidations)
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
	blocksPool *mocks.BlocksPool,
	logToConsole bool,
	withFailingBlockProposalValidations bool,
) *Node {

	communicationInstance := mocks.NewCommunication(memberId, discovery)
	if tb.communicationMaxDelay > time.Duration(0) {
		communicationInstance.SetMessagesMaxDelay(tb.communicationMaxDelay)
	}
	discovery.RegisterCommunication(memberId, communicationInstance)
	membership := mocks.NewMockMembership(memberId, discovery, tb.orderCommitteeByHeight)

	blockUtils := mocks.NewMockBlockUtils(memberId, blocksPool)
	if withFailingBlockProposalValidations {
		blockUtils = blockUtils.WithFailingBlockProposalValidations()
	}

	b := nodeBuilder.
		AsInstanceId(tb.instanceId).
		CommunicatesVia(communicationInstance).
		ThatIsPartOf(membership).
		WithBlockUtils(blockUtils).
		WithMemberId(memberId)

	//if tb.blockUtils != nil {
	//	b.WithBlockUtils(tb.blockUtils)
	//}

	if logToConsole {
		b.ThatLogsToConsole()
	}

	if tb.useTimeBasedElectionTrigger {
		et := electiontrigger.NewTimerBasedElectionTrigger(tb.electionTriggerTimeout, nil)
		b.WithElectionTrigger(et)
	}
	return b.Build()
}

func (tb *TestNetworkBuilder) createNodes(discovery *mocks.Discovery, blocksPool *mocks.BlocksPool, logToConsole bool, withFailingBlockProposalValidations bool) []*Node {
	var nodes []*Node
	for i := 0; i < tb.NodeCount; i++ {
		nodeBuilder := NewNodeBuilder()
		memberId := primitives.MemberId(fmt.Sprintf("%03d", i))
		node := tb.buildNode(nodeBuilder, memberId, discovery, blocksPool, logToConsole, withFailingBlockProposalValidations)
		nodes = append(nodes, node)
	}

	for i, customBuilder := range tb.customNodeBuilders {
		memberId := primitives.MemberId(fmt.Sprintf("C02%d", i))
		node := tb.buildNode(customBuilder, memberId, discovery, blocksPool, logToConsole, withFailingBlockProposalValidations)
		nodes = append(nodes, node)
	}

	return nodes
}

func (tb *TestNetworkBuilder) WithCommunication(communication interfaces.Communication) *TestNetworkBuilder {
	tb.communication = communication
	return tb
}

func (tb *TestNetworkBuilder) WithBlockUtils(utils interfaces.BlockUtils) *TestNetworkBuilder {
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
		NodeCount:          0,
		customNodeBuilders: nil,
		upcomingBlocks:     nil,
	}
}

func ABasicTestNetwork(ctx context.Context) *TestNetwork {
	return ATestNetworkBuilder(4).Build(ctx)
}

func ABasicTestNetworkWithConsoleLogs(ctx context.Context) *TestNetwork {
	return ATestNetworkBuilder(4).LogToConsole().Build(ctx)
}

func ATestNetworkBuilder(countOfNodes int, blocksPool ...interfaces.Block) *TestNetworkBuilder {
	instanceId := primitives.InstanceId(rand.Uint64())
	return NewTestNetworkBuilder().
		WithNodeCount(countOfNodes).
		WithBlocks(blocksPool...).
		InNetwork(instanceId)
	//LogToConsole().
}
