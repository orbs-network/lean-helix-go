package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoViewChangeMessages(t *testing.T) {
	actual := leanhelix.GetLatestBlockFromViewChangeMessages([]*leanhelix.ViewChangeMessage{})
	require.Nil(t, actual, "Should have returned Nil for no ViewChange Messages")
}

func TestReturnNilWhenNoViewChangeMessages(t *testing.T) {
	keyManager := builders.NewMockKeyManager(primitives.Ed25519PublicKey("PublicKey 1"))
	messageFactory := leanhelix.NewMessageFactory(keyManager)
	VCMessage := messageFactory.CreateViewChangeMessage(1, 2, nil)

	actual := leanhelix.GetLatestBlockFromViewChangeMessages([]*leanhelix.ViewChangeMessage{VCMessage})
	require.Nil(t, actual, "Should have returned Nil for ViewChange Messages without prepared messages")
}

func TestKeepOnlyMessagesWithBlock(t *testing.T) {
	keyManager1 := builders.NewMockKeyManager(primitives.Ed25519PublicKey("PublicKey 1"))
	keyManager2 := builders.NewMockKeyManager(primitives.Ed25519PublicKey("PublicKey 2"))
	keyManager3 := builders.NewMockKeyManager(primitives.Ed25519PublicKey("PublicKey 3"))

	messageFactory1 := leanhelix.NewMessageFactory(keyManager1)
	messageFactory2 := leanhelix.NewMessageFactory(keyManager2)
	messageFactory3 := leanhelix.NewMessageFactory(keyManager3)

	block := builders.CreateBlock(builders.GenesisBlock)
	PMessage1 := messageFactory1.CreatePrepareMessage(1, 2, block.BlockHash())
	PMessage2 := messageFactory2.CreatePrepareMessage(1, 2, block.BlockHash())

	preparedMessages := &leanhelix.PreparedMessages{
		PreprepareMessage: nil,
		PrepareMessages: []*leanhelix.PrepareMessage{
			PMessage1,
			PMessage2,
		},
	}

	VCMessage := messageFactory3.CreateViewChangeMessage(1, 2, preparedMessages)

	actual := leanhelix.GetLatestBlockFromViewChangeMessages([]*leanhelix.ViewChangeMessage{VCMessage})
	require.Nil(t, actual, "A block returned from View Change messages without block")
}

func TestReturnBlockFromPPMWithHighestView(t *testing.T) {
	testNetwork := builders.ABasicTestNetwork()
	node0 := testNetwork.Nodes[0]
	node1 := testNetwork.Nodes[1]
	node2 := testNetwork.Nodes[2]
	node3 := testNetwork.Nodes[3]

	mf0 := leanhelix.NewMessageFactory(node0.KeyManager)
	mf2 := leanhelix.NewMessageFactory(node2.KeyManager)

	// view on view 3
	blockOnView3 := builders.CreateBlock(builders.GenesisBlock)
	preparedOnView3 := builders.CreatePreparedMessages(node3, []*builders.Node{node1, node2}, 1, 3, blockOnView3)

	VCMessageOnView3 := mf0.CreateViewChangeMessage(1, 5, preparedOnView3)

	// view on view 8
	blockOnView8 := builders.CreateBlock(builders.GenesisBlock)
	preparedOnView8 := builders.CreatePreparedMessages(node0, []*builders.Node{node1, node2}, 1, 8, blockOnView8)
	VCMessageOnView8 := mf2.CreateViewChangeMessage(1, 5, preparedOnView8)

	// view on view 4
	blockOnView4 := builders.CreateBlock(builders.GenesisBlock)
	preparedOnView4 := builders.CreatePreparedMessages(node0, []*builders.Node{node1, node2}, 1, 4, blockOnView4)
	VCMessageOnView4 := mf2.CreateViewChangeMessage(1, 5, preparedOnView4)

	actual := leanhelix.GetLatestBlockFromViewChangeMessages([]*leanhelix.ViewChangeMessage{VCMessageOnView3, VCMessageOnView8, VCMessageOnView4})
	require.Equal(t, blockOnView8, actual, "Returned block is not from the latest view")
}
