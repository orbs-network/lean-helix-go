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
	memberId        primitives.MemberId
	electionTrigger interfaces.ElectionScheduler
	blockUtils      interfaces.BlockUtils
	l               interfaces.Logger
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

func (builder *NodeBuilder) WithElectionTrigger(electionTrigger interfaces.ElectionScheduler) *NodeBuilder {
	if builder.electionTrigger == nil {
		builder.electionTrigger = electionTrigger
	}
	return builder
}

func (builder *NodeBuilder) WithLogger(logger interfaces.Logger) *NodeBuilder {
	builder.l = logger
	return builder
}

func (builder *NodeBuilder) WithBlockUtils(utils interfaces.BlockUtils) *NodeBuilder {
	builder.blockUtils = utils
	return builder
}

func (builder *NodeBuilder) Build() *Node {
	memberId := builder.memberId
	if memberId == nil {
		memberId = primitives.MemberId(fmt.Sprintf("XXX"))
	}

	if builder.l == nil {
		builder.l = logger.NewSilentLogger()
	}

	if builder.blockUtils == nil {
		builder.blockUtils = mocks.NewMockBlockUtils(builder.memberId, mocks.NewBlocksPool(nil), builder.l)
	}

	return NewNode(
		builder.instanceId,
		builder.membership,
		builder.communication,
		builder.blockUtils,
		builder.electionTrigger,
		builder.l,
	)
}

func ADummyNode() *Node {
	memberId := primitives.MemberId("Dummy")
	return NewNodeBuilder().
		WithMemberId(memberId).
		ThatIsPartOf(mocks.NewFakeMembership(memberId, nil, nil, false)).
		CommunicatesVia(mocks.NewCommunication(memberId, nil, nil)).
		Build()
}
