package byzantineattacks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
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
			WithBlocks([]interfaces.Block{}).
			WithTimeBasedElectionTrigger(time.Duration(200) * time.Millisecond).
			GossipMessagesMaxDelay(time.Duration(100) * time.Millisecond).
			WithNodeCount(4).
			//LogToConsole().
			Build()

		net.Nodes[0].WriteToStateChannel = false
		net.Nodes[1].WriteToStateChannel = false
		net.Nodes[2].WriteToStateChannel = false
		net.Nodes[3].WriteToStateChannel = false

		net.StartConsensus(ctx)

		time.Sleep(time.Duration(5) * time.Second)
		// todo add a watch to the nods blockchain, and wait for 100 blocks
	})
}
