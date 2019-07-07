// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package network

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

type NodeBuilder struct {
	instanceId      primitives.InstanceId
	communication   *mocks.CommunicationMock
	membership      interfaces.Membership
	blocksPool      *mocks.BlocksPool
	logsToConsole   bool
	memberId        primitives.MemberId
	electionTrigger interfaces.ElectionTrigger
	blockUtils      mocks.MockBlockUtils
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

func (builder *NodeBuilder) CommunicatesVia(communication *mocks.CommunicationMock) *NodeBuilder {
	if builder.communication == nil {
		builder.communication = communication
	}
	return builder
}

func (builder *NodeBuilder) ThatIsPartOf(membership interfaces.Membership) *NodeBuilder {
	if builder.membership == nil {
		builder.membership = membership
	}
	return builder
}

func (builder *NodeBuilder) WithMemberId(memberId primitives.MemberId) *NodeBuilder {
	if builder.memberId == nil {
		builder.memberId = memberId
	}
	return builder
}

func (builder *NodeBuilder) AsInstanceId(instanceId primitives.InstanceId) *NodeBuilder {
	builder.instanceId = instanceId
	return builder
}

func (builder *NodeBuilder) WithBlocksPool(blocksPool *mocks.BlocksPool) *NodeBuilder {
	if builder.blocksPool == nil {
		builder.blocksPool = blocksPool
	}
	return builder
}

func (builder *NodeBuilder) WithElectionTrigger(electionTrigger interfaces.ElectionTrigger) *NodeBuilder {
	if builder.electionTrigger == nil {
		builder.electionTrigger = electionTrigger
	}
	return builder
}

func (builder *NodeBuilder) ThatLogsToConsole() *NodeBuilder {
	builder.logsToConsole = true
	return builder
}

func (builder *NodeBuilder) WithBlockUtils(utils mocks.MockBlockUtils) {
	builder.blockUtils = utils
}

func (builder *NodeBuilder) Build() *Node {
	memberId := builder.memberId
	if memberId == nil {
		memberId = primitives.MemberId(fmt.Sprintf("Dummy MemberId"))
	}

	if builder.blockUtils == nil {
		builder.blockUtils = mocks.NewMockBlockUtils(builder.blocksPool)
	}

	var electionTrigger interfaces.ElectionTrigger
	if builder.electionTrigger == nil {
		electionTrigger = mocks.NewMockElectionTrigger()
	} else {
		electionTrigger = builder.electionTrigger
	}

	var l interfaces.Logger
	if builder.logsToConsole {
		l = logger.NewConsoleLogger()
	}
	return NewNode(builder.instanceId, builder.membership, builder.communication, builder.blockUtils, electionTrigger, l)
}

func ADummyNode() *Node {
	memberId := primitives.MemberId("Dummy")
	return NewNodeBuilder().
		WithMemberId(memberId).
		ThatIsPartOf(mocks.NewMockMembership(memberId, nil, false)).
		CommunicatesVia(mocks.NewCommunication(nil)).
		Build()
}
