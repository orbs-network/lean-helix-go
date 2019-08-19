// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package byzantineattacks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/network"
	"math/rand"
	"testing"
	"time"
)

func TestThatWeReachConsensusEventIfWeDelayAllTheGossipMessages(t *testing.T) {
	rand.Seed(time.Now().Unix())
	test.WithContext(func(ctx context.Context) {
		net := network.
			NewTestNetworkBuilder().
			WithBlocks().
			WithTimeBasedElectionTrigger(1000 * time.Millisecond).
			GossipMessagesMaxDelay(100 * time.Millisecond).
			WithNodeCount(4).
			//LogToConsole().
			Build(ctx)

		net.Nodes[0].WriteToStateChannel = false
		net.Nodes[1].WriteToStateChannel = false
		net.Nodes[2].WriteToStateChannel = false
		net.Nodes[3].WriteToStateChannel = false

		net.StartConsensus(ctx)

		time.Sleep(1 * time.Second)
		// todo add a watch to the nods blockchain, and wait for 100 blocks
	})
}
