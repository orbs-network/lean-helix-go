package test

import (
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageBuilderAndReader(t *testing.T) {
	height := BlockHeight(math.Floor(rand.Float64() * 1000000000))
	view := View(math.Floor(rand.Float64() * 1000000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	b1 := builders.CreateBlock(builders.GenesisBlock)
	mockKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK"), nil)
	mf := &lh.MessageFactory{
		KeyManager: mockKeyManager,
	}

	t.Run("build and read PreprepareMessage", func(t *testing.T) {
		ppm := mf.CreatePreprepareMessage(height, view, b1)
		ppmBytes := ppm.Raw()
		receivedPPMC := lh.PreprepareContentReader(ppmBytes)
		require.Equal(t, receivedPPMC.SignedHeader().MessageType(), lh.LEAN_HELIX_PREPREPARE, "Message type should be LEAN_HELIX_PREPREPARE")
		require.True(t, receivedPPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read PrepareMessage", func(t *testing.T) {
		pm := mf.CreatePrepareMessage(height, view, b1.BlockHash())
		pmBytes := pm.Raw()
		receivedPMC := lh.PrepareContentReader(pmBytes)
		require.Equal(t, receivedPMC.SignedHeader().MessageType(), lh.LEAN_HELIX_PREPARE, "Message type should be LEAN_HELIX_PREPARE")
		require.True(t, receivedPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read CommitMessage", func(t *testing.T) {
		cm := mf.CreateCommitMessage(height, view, b1.BlockHash())
		cmBytes := cm.Raw()
		receivedCMC := lh.CommitContentReader(cmBytes)
		require.Equal(t, receivedCMC.SignedHeader().MessageType(), lh.LEAN_HELIX_COMMIT, "Message type should be LEAN_HELIX_COMMIT")
		require.True(t, receivedCMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedCMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read ViewChangeMessage", func(t *testing.T) {
		vcm := mf.CreateViewChangeMessage(height, view, nil)
		vcmBytes := vcm.Raw()
		receivedVCMC := lh.ViewChangeMessageContentReader(vcmBytes)
		require.Equal(t, receivedVCMC.SignedHeader().MessageType(), lh.LEAN_HELIX_VIEW_CHANGE, "Message type should be LEAN_HELIX_VIEW_CHANGE")
		require.True(t, receivedVCMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedVCMC.SignedHeader().View().Equal(view), "View = %v", view)
		// TODO add test with preparedproof and do require.Equal on some of the proof's internal properties
	})
	t.Run("build and read NewViewMessage", func(t *testing.T) {
		ppmcb := mf.CreatePreprepareMessageContentBuilder(height, view, b1)
		vcm1 := mf.CreateViewChangeMessageContentBuilder(height, view, nil)
		vcm2 := mf.CreateViewChangeMessageContentBuilder(height, view, nil)
		confirmation1 := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: vcm1.SignedHeader,
			Sender:       vcm1.Sender,
		}

		confirmation2 := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: vcm2.SignedHeader,
			Sender:       vcm2.Sender,
		}
		nvm := mf.CreateNewViewMessage(height, view, ppmcb, []*lh.ViewChangeMessageContentBuilder{confirmation1, confirmation2}, block)
		nvmBytes := nvm.Raw()
		receivedNVMC := lh.NewViewMessageContentReader(nvmBytes)
		require.Equal(t, receivedNVMC.SignedHeader().MessageType(), lh.LEAN_HELIX_NEW_VIEW, "Message type should be LEAN_HELIX_NEW_VIEW")
		require.True(t, receivedNVMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedNVMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
}
