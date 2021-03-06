// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestMessageBuilderAndReader(t *testing.T) {
	instanceId := primitives.InstanceId(rand.Uint64())
	height := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	block := mocks.ABlock(interfaces.GenesisBlock)
	b1 := mocks.ABlock(interfaces.GenesisBlock)
	memberId := primitives.MemberId("Member Id")
	mockKeyManager := mocks.NewMockKeyManager(memberId, nil)
	mf := messagesfactory.NewMessageFactory(instanceId, mockKeyManager, memberId, 0)

	t.Run("build and read PreprepareMessage", func(t *testing.T) {
		ppm := mf.CreatePreprepareMessage(height, view, b1, mocks.CalculateBlockHash(b1))
		ppmBytes := ppm.Raw()
		receivedPPMC := protocol.PreprepareContentReader(ppmBytes)
		require.Equal(t, receivedPPMC.SignedHeader().MessageType(), protocol.LEAN_HELIX_PREPREPARE, "Message type should be LEAN_HELIX_PREPREPARE")
		require.True(t, receivedPPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read PrepareMessage", func(t *testing.T) {
		pm := mf.CreatePrepareMessage(height, view, mocks.CalculateBlockHash(b1))
		pmBytes := pm.Raw()
		receivedPMC := protocol.PrepareContentReader(pmBytes)
		require.Equal(t, receivedPMC.SignedHeader().MessageType(), protocol.LEAN_HELIX_PREPARE, "Message type should be LEAN_HELIX_PREPARE")
		require.True(t, receivedPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read CommitMessage", func(t *testing.T) {
		cm := mf.CreateCommitMessage(height, view, mocks.CalculateBlockHash(b1))
		cmBytes := cm.Raw()
		receivedCMC := protocol.CommitContentReader(cmBytes)
		require.Equal(t, receivedCMC.SignedHeader().MessageType(), protocol.LEAN_HELIX_COMMIT, "Message type should be LEAN_HELIX_COMMIT")
		require.True(t, receivedCMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedCMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read ViewChangeMessage", func(t *testing.T) {
		vcm := mf.CreateViewChangeMessage(height, view, nil)
		vcmBytes := vcm.Raw()
		receivedVCMC := protocol.ViewChangeMessageContentReader(vcmBytes)
		require.Equal(t, receivedVCMC.SignedHeader().MessageType(), protocol.LEAN_HELIX_VIEW_CHANGE, "Message type should be LEAN_HELIX_VIEW_CHANGE")
		require.True(t, receivedVCMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedVCMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read NewViewMessage", func(t *testing.T) {
		ppmcb := mf.CreatePreprepareMessageContentBuilder(height, view, b1, mocks.CalculateBlockHash(b1))
		vcm1 := mf.CreateViewChangeMessageContentBuilder(height, view, nil)
		vcm2 := mf.CreateViewChangeMessageContentBuilder(height, view, nil)
		confirmation1 := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: vcm1.SignedHeader,
			Sender:       vcm1.Sender,
		}

		confirmation2 := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: vcm2.SignedHeader,
			Sender:       vcm2.Sender,
		}
		nvm := mf.CreateNewViewMessage(height, view, ppmcb, []*protocol.ViewChangeMessageContentBuilder{confirmation1, confirmation2}, block)
		nvmBytes := nvm.Raw()
		receivedNVMC := protocol.NewViewMessageContentReader(nvmBytes)
		require.Equal(t, receivedNVMC.SignedHeader().MessageType(), protocol.LEAN_HELIX_NEW_VIEW, "Message type should be LEAN_HELIX_NEW_VIEW")
		require.True(t, receivedNVMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedNVMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
}
