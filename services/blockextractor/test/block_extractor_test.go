package test

import (
	"github.com/orbs-network/lean-helix-go/services/blockextractor"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoViewChangeMessages(t *testing.T) {
	actual := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{})
	require.Nil(t, actual, "Should have returned Nil for no ViewChange Messages")
}

func TestReturnNilWhenNoViewChangeMessages(t *testing.T) {
	memberId := primitives.MemberId("MemberId 1")
	keyManager := mocks.NewMockKeyManager(memberId)
	VCMessage := builders.AViewChangeMessage(keyManager, memberId, 1, 2, nil)

	actual := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{VCMessage})
	require.Nil(t, actual, "Should have returned Nil for ViewChange Messages without prepared messages")
}

func TestKeepOnlyMessagesWithBlock(t *testing.T) {
	memberId1 := primitives.MemberId("MemberId 1")
	memberId2 := primitives.MemberId("MemberId 2")
	memberId3 := primitives.MemberId("MemberId 3")
	keyManager1 := mocks.NewMockKeyManager(memberId1)
	keyManager2 := mocks.NewMockKeyManager(memberId2)
	keyManager3 := mocks.NewMockKeyManager(memberId3)

	block := mocks.CreateBlock(interfaces.GenesisBlock)

	preparedMessages := &preparedmessages.PreparedMessages{
		PreprepareMessage: nil,
		PrepareMessages: []*interfaces.PrepareMessage{
			builders.APrepareMessage(keyManager1, memberId1, 1, 2, block),
			builders.APrepareMessage(keyManager2, memberId2, 1, 2, block),
		},
	}

	VCMessage := builders.AViewChangeMessage(keyManager3, memberId3, 1, 2, preparedMessages)

	actual := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{VCMessage})
	require.Nil(t, actual, "A block returned from View Change messages without block")
}

func TestReturnBlockFromPPMWithHighestView(t *testing.T) {
	testNetwork := network.ABasicTestNetwork()
	node0 := testNetwork.Nodes[0]
	node1 := testNetwork.Nodes[1]
	node2 := testNetwork.Nodes[2]
	node3 := testNetwork.Nodes[3]

	// view on view 3
	blockOnView3 := mocks.CreateBlock(interfaces.GenesisBlock)
	preparedOnView3 := builders.CreatePreparedMessages(node3, []builders.Sender{node1, node2}, 1, 3, blockOnView3)

	VCMessageOnView3 := builders.AViewChangeMessage(node0.KeyManager, node0.MemberId, 1, 5, preparedOnView3)

	// view on view 8
	blockOnView8 := mocks.CreateBlock(interfaces.GenesisBlock)
	preparedOnView8 := builders.CreatePreparedMessages(node0, []builders.Sender{node1, node2}, 1, 8, blockOnView8)
	VCMessageOnView8 := builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 5, preparedOnView8)

	// view on view 4
	blockOnView4 := mocks.CreateBlock(interfaces.GenesisBlock)
	preparedOnView4 := builders.CreatePreparedMessages(node0, []builders.Sender{node1, node2}, 1, 4, blockOnView4)
	VCMessageOnView4 := builders.AViewChangeMessage(node2.KeyManager, node2.MemberId, 1, 5, preparedOnView4)

	actual := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{VCMessageOnView3, VCMessageOnView8, VCMessageOnView4})
	require.Equal(t, blockOnView8, actual, "Returned block is not from the latest view")
}
