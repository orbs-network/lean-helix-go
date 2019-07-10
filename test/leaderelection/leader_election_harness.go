// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"testing"
)

type harness struct {
	t   *testing.T
	net *network.TestNetwork
}

func NewHarness(ctx context.Context, t *testing.T, blocksPool ...interfaces.Block) *harness {
	return newHarness(ctx, t, mocks.NewMockBlockUtils(mocks.NewBlocksPool(blocksPool)))
}

func NewHarnessWithFailingBlockProposalValidations(ctx context.Context, t *testing.T) *harness {
	//net := builders.NewTestNetworkBuilder().WithNodeCount(4).WithBlocks(blocksPool).LogToConsole().Build()
	failingValidationBlocksUtils := mocks.NewMockBlockUtils(mocks.NewBlocksPool(nil)).WithFailingBlockProposalValidations()
	return newHarness(ctx, t, failingValidationBlocksUtils)
}

func newHarness(ctx context.Context, t *testing.T, blockUtils interfaces.BlockUtils) *harness {
	//net := builders.NewTestNetworkBuilder().WithNodeCount(4).WithBlocks(blocksPool).LogToConsole().Build()
	networkBuilder := network.ATestNetworkBuilder(4)
	if blockUtils != nil {
		networkBuilder = networkBuilder.WithBlockUtils(blockUtils)
	}
	net := networkBuilder.Build()
	net.SetNodesToPauseOnRequestNewBlock()
	net.StartConsensus(ctx)

	return &harness{
		t:   t,
		net: net,
	}
}

func (h *harness) TriggerElectionOnAllNodes(ctx context.Context) {
	h.net.TriggerElectionOnAllNodes(ctx)
}
