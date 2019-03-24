// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

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
	"math/rand"
	"testing"
)

func TestNoViewChangeMessages(t *testing.T) {
	actualBlock, actualBlockHash := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{})
	require.Nil(t, actualBlock, "Should have returned Nil for no ViewChange Messages")
	require.Nil(t, actualBlockHash, "Should have returned Nil for no ViewChange Messages")
}

func TestReturnNilWhenNoViewChangeMessages(t *testing.T) {
	memberId := primitives.MemberId("MemberId 1")
	keyManager := mocks.NewMockKeyManager(memberId)
	instanceId := primitives.InstanceId(rand.Uint64())
	VCMessage := builders.AViewChangeMessage(instanceId, keyManager, memberId, 1, 2, nil)

	actualBlock, actualBlockHash := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{VCMessage})
	require.Nil(t, actualBlock, "Should have returned Nil for ViewChange Messages without prepared messages")
	require.Nil(t, actualBlockHash, "Should have returned Nil for ViewChange Messages without prepared messages")
}

func TestKeepOnlyMessagesWithBlock(t *testing.T) {
	instanceId := primitives.InstanceId(rand.Uint64())
	memberId1 := primitives.MemberId("MemberId 1")
	memberId2 := primitives.MemberId("MemberId 2")
	memberId3 := primitives.MemberId("MemberId 3")
	keyManager1 := mocks.NewMockKeyManager(memberId1)
	keyManager2 := mocks.NewMockKeyManager(memberId2)
	keyManager3 := mocks.NewMockKeyManager(memberId3)

	block := mocks.ABlock(interfaces.GenesisBlock)

	preparedMessages := &preparedmessages.PreparedMessages{
		PreprepareMessage: nil,
		PrepareMessages: []*interfaces.PrepareMessage{
			builders.APrepareMessage(instanceId, keyManager1, memberId1, 1, 2, block),
			builders.APrepareMessage(instanceId, keyManager2, memberId2, 1, 2, block),
		},
	}

	VCMessage := builders.AViewChangeMessage(instanceId, keyManager3, memberId3, 1, 2, preparedMessages)

	actualBlock, actualBlockHash := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{VCMessage})
	require.Nil(t, actualBlock, "A block returned from View Change messages without block")
	require.Nil(t, actualBlockHash, "A block returned from View Change messages without block")
}

func TestReturnBlockFromPPMWithHighestView(t *testing.T) {
	instanceId := primitives.InstanceId(rand.Uint64())
	testNetwork := network.ABasicTestNetwork()
	node0 := testNetwork.Nodes[0]
	node1 := testNetwork.Nodes[1]
	node2 := testNetwork.Nodes[2]
	node3 := testNetwork.Nodes[3]

	// view on view 3
	blockOnView3 := mocks.ABlock(interfaces.GenesisBlock)
	preparedOnView3 := builders.CreatePreparedMessages(instanceId, node3, []builders.Sender{node1, node2}, 1, 3, blockOnView3)

	VCMessageOnView3 := builders.AViewChangeMessage(instanceId, node0.KeyManager, node0.MemberId, 1, 5, preparedOnView3)

	// view on view 8
	blockOnView8 := mocks.ABlock(interfaces.GenesisBlock)
	blockHashOnView8 := mocks.CalculateBlockHash(blockOnView8)
	preparedOnView8 := builders.CreatePreparedMessages(instanceId, node0, []builders.Sender{node1, node2}, 1, 8, blockOnView8)
	VCMessageOnView8 := builders.AViewChangeMessage(instanceId, node2.KeyManager, node2.MemberId, 1, 5, preparedOnView8)

	// view on view 4
	blockOnView4 := mocks.ABlock(interfaces.GenesisBlock)
	preparedOnView4 := builders.CreatePreparedMessages(instanceId, node0, []builders.Sender{node1, node2}, 1, 4, blockOnView4)
	VCMessageOnView4 := builders.AViewChangeMessage(instanceId, node2.KeyManager, node2.MemberId, 1, 5, preparedOnView4)

	actualBlock, actualBlockHash := blockextractor.GetLatestBlockFromViewChangeMessages([]*interfaces.ViewChangeMessage{VCMessageOnView3, VCMessageOnView8, VCMessageOnView4})
	require.Equal(t, blockOnView8, actualBlock, "Returned block is not from the latest view")
	require.Equal(t, blockHashOnView8, actualBlockHash, "Returned block is not from the latest view")
}
