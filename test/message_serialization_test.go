package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageBuilderAndReader(t *testing.T) {
	height := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	b1 := builders.CreateBlock(builders.GenesisBlock)
	mockKeyManager := builders.NewMockKeyManager(primitives.MemberId("Member Id"), nil)
	mf := &leanhelix.MessageFactory{
		KeyManager: mockKeyManager,
	}

	t.Run("build and read PreprepareMessage", func(t *testing.T) {
		ppm := mf.CreatePreprepareMessage(height, view, b1, builders.CalculateBlockHash(b1))
		ppmBytes := ppm.Raw()
		receivedPPMC := protocol.PreprepareContentReader(ppmBytes)
		require.Equal(t, receivedPPMC.SignedHeader().MessageType(), protocol.LEAN_HELIX_PREPREPARE, "Message type should be LEAN_HELIX_PREPREPARE")
		require.True(t, receivedPPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read PrepareMessage", func(t *testing.T) {
		pm := mf.CreatePrepareMessage(height, view, builders.CalculateBlockHash(b1))
		pmBytes := pm.Raw()
		receivedPMC := protocol.PrepareContentReader(pmBytes)
		require.Equal(t, receivedPMC.SignedHeader().MessageType(), protocol.LEAN_HELIX_PREPARE, "Message type should be LEAN_HELIX_PREPARE")
		require.True(t, receivedPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read CommitMessage", func(t *testing.T) {
		cm := mf.CreateCommitMessage(height, view, builders.CalculateBlockHash(b1))
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
		// TODO add test with preparedproof and do require.Equal on some of the proof's internal properties
	})
	t.Run("build and read NewViewMessage", func(t *testing.T) {
		ppmcb := mf.CreatePreprepareMessageContentBuilder(height, view, b1, builders.CalculateBlockHash(b1))
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
