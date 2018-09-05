package leaderelection

import (
	"testing"
)

const nodeCount = 5

func TestLeaderElection(t *testing.T) {

	t.Run("should notify the next leader when the timeout expired", func(t *testing.T) {
		//h := NewHarness(nodeCount)

		// Impl me
	})

	t.Run("should cycle back to the first node on view-change", func(t *testing.T) {
		// Impl me
	})

	t.Run("should count 2f+1 view-change to be elected", func(t *testing.T) {
		// impl me
		// uses gossip.onRemoteMessage

		//block1 := builders.CreateBlock(builders.GenesisBlock)
		//block2 := builders.CreateBlock(block1)
		//testNetwork := CreateSimpleTestNetwork(4, {block1, block2})
		//
		//node0 := testNetwork.Nodes[0]
		//node1 := testNetwork.Nodes[1]
		//node2 := testNetwork.Nodes[2]
		//node3 := testNetwork.Nodes[3]
		//gossip := testNetwork.GetNodeGossip(node1.PublicKey)
		//...

	})

}
