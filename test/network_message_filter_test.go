package test

import (
	"context"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"testing"
)

func TestSetBlockHeightAndReceiveGossipMessages(t *testing.T) {
	ctx := context.Background()

	height := primitives.BlockHeight(3)
	view := primitives.View(0)
	net := builders.NewSimpleTestNetwork(NODE_COUNT, nil) // Node 0 is leader

	node0 := net.Nodes[0]
	node1 := net.Nodes[1]
	node1MessageFactory := leanhelix.NewMessageFactory(node1.KeyManager)
	mockReceiver := builders.NewMockMessageReceiver()

	mockReceiver.When("OnReceivePreprepare", mock.Any, mock.Any).Times(1)
	mockReceiver.When("OnReceivePrepare", mock.Any, mock.Any).Times(1)
	mockReceiver.When("OnReceiveCommit", mock.Any, mock.Any).Times(1)
	mockReceiver.When("OnReceiveViewChange", mock.Any, mock.Any).Times(1)
	mockReceiver.When("OnReceiveNewView", mock.Any, mock.Any).Times(1)

	gossip := net.GetNodeGossip(node1.KeyManager.MyPublicKey())
	gossip.When("SendMessage", mock.Any, mock.Any).Times(5)
	filter := leanhelix.NewNetworkMessageFilter(gossip, height, node0.KeyManager.MyPublicKey(), mockReceiver)
	filter.SetBlockHeight(ctx, 3)
	block := builders.CreateBlock(builders.GenesisBlock)
	allNetworkPublicKeys := net.Discovery.AllGossipsPublicKeys()

	messages := make([]leanhelix.ConsensusMessage, 5)

	messages[0] = node1MessageFactory.CreatePreprepareMessage(height, view, block)
	messages[1] = node1MessageFactory.CreatePrepareMessage(height, view, block.BlockHash())
	messages[2] = node1MessageFactory.CreateCommitMessage(height, view, block.BlockHash())
	messages[3] = node1MessageFactory.CreateViewChangeMessage(height, view, nil)
	messages[4] = node1MessageFactory.CreateNewViewMessage(height, view, nil, nil, block)

	for _, msg := range messages {
		rawMsg := msg.ToConsensusRawMessage()
		if err := gossip.SendMessage(ctx, allNetworkPublicKeys, rawMsg); err != nil {
			t.Error(err)
		}
	}

	ok, err := mockReceiver.Verify()
	if !ok {
		t.Error(err)
	}

}
