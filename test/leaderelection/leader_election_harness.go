// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/network"
	"testing"
)

type harness struct {
	t   *testing.T
	net *network.TestNetwork
}

func NewStartedHarness(ctx context.Context, t *testing.T, logsToConsole bool, blocksPool ...interfaces.Block) *harness {
	return newHarness(ctx, t, logsToConsole, false, true, blocksPool...)
}

func NewStartedHarnessDontPauseOnRequestNewBlock(ctx context.Context, t *testing.T, logsToConsole bool, blocksPool ...interfaces.Block) *harness {
	return newHarness(ctx, t, logsToConsole, false, false, blocksPool...)
}

// This might not be a good idea but it is needed outside this package
func Net(h *harness) *network.TestNetwork {
	return h.net
}
func NewStartedHarnessWithFailingBlockProposalValidations(ctx context.Context, t *testing.T, logsToConsole bool) *harness {
	return newHarness(ctx, t, logsToConsole, true, true)
}

func newHarness(ctx context.Context, t *testing.T, logsToConsole bool, withFailingBlockProposalValidations bool, pauseOnRequestNewBlock bool, blocksPool ...interfaces.Block) *harness {
	networkBuilder := network.ATestNetworkBuilder(4)
	if logsToConsole {
		networkBuilder = networkBuilder.LogToConsole(t)
	}
	net := networkBuilder.
		WithMaybeFailingBlockProposalValidations(withFailingBlockProposalValidations, blocksPool...).
		Build(ctx)

	// Create all channels in advance, using the test context which will only get canceled at the end of test
	for _, node := range net.Nodes {
		for _, peerNode := range net.Nodes {
			net.GetNodeCommunication(node.MemberId).ReturnAndMaybeCreateOutgoingChannelByTarget(ctx, peerNode.MemberId)
		}
	}

	if pauseOnRequestNewBlock {
		net.SetNodesToPauseOnRequestNewBlock()
	} else {
		net.SetNodesToNotPauseOnRequestNewBlock()
	}
	net.StartConsensus(ctx)

	return &harness{
		t:   t,
		net: net,
	}
}
