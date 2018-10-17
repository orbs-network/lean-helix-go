package builders

import (
	"bytes"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestBuildAndRead(t *testing.T) {
	keyManager := NewMockKeyManager(Ed25519PublicKey("My PK"))
	height := BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := View(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(GenesisBlock)
	fac := lh.NewMessageFactory(keyManager)

	actualPPM := fac.CreatePreprepareMessage(height, view, block)

	bytes1 := actualPPM.Raw()
	newPPMC := lh.PreprepareMessageContentReader(bytes1)
	bytes2 := newPPMC.Raw()

	require.True(t, bytes.Compare(bytes1, bytes2) == 0)
}

func TestMessageFactory(t *testing.T) {
	leaderKeyManager := NewMockKeyManager(Ed25519PublicKey("PK0"))
	node1KeyManager := NewMockKeyManager(Ed25519PublicKey("PK1"))
	node2KeyManager := NewMockKeyManager(Ed25519PublicKey("PK2"))
	height := BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := View(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(GenesisBlock)
	blockHash := block.BlockHash()
	leaderFac := lh.NewMessageFactory(leaderKeyManager)
	node1Fac := lh.NewMessageFactory(node1KeyManager)
	node2Fac := lh.NewMessageFactory(node2KeyManager)
	b1 := CreateBlock(GenesisBlock)
	mockKeyManager := NewMockKeyManager(Ed25519PublicKey("PK"), nil)
	mf := &lh.MessageFactoryImpl{
		KeyManager: mockKeyManager,
	}

	t.Run("build and read PreprepareMessage", func(t *testing.T) {
		ppm := mf.CreatePreprepareMessage(height, view, b1)
		ppmBytes := ppm.Raw()
		receivedPPMC := lh.PreprepareMessageContentReader(ppmBytes)
		require.Equal(t, receivedPPMC.SignedHeader().MessageType(), lh.LEAN_HELIX_PREPREPARE, "Message type should be LEAN_HELIX_PREPREPARE")
		require.True(t, receivedPPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read PrepareMessage", func(t *testing.T) {
		pm := mf.CreatePrepareMessage(height, view, b1.BlockHash())
		pmBytes := pm.Raw()
		receivedPMC := lh.PrepareMessageContentReader(pmBytes)
		require.Equal(t, receivedPMC.SignedHeader().MessageType(), lh.LEAN_HELIX_PREPARE, "Message type should be LEAN_HELIX_PREPARE")
		require.True(t, receivedPMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedPMC.SignedHeader().View().Equal(view), "View = %v", view)
	})
	t.Run("build and read CommitMessage", func(t *testing.T) {
		cm := mf.CreateCommitMessage(height, view, b1.BlockHash())
		cmBytes := cm.Raw()
		receivedCMC := lh.CommitMessageContentReader(cmBytes)
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
		ppm := mf.CreatePreprepareMessage(height, view, b1)
		vcm1 := mf.CreateViewChangeMessageContentBuilder(height, view, nil)
		vcm2 := mf.CreateViewChangeMessageContentBuilder(height, view, nil)
		vcContent := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: vcm1.SignedHeader,
			Sender:       vcm1.Sender,
		}

		confirmationBuilder := &lh.ViewChangeConfirmationBuilder{}
		nvm := mf.CreateNewViewMessage(height, view, ppm, confirmations)
		nvmBytes := nvm.Raw()
		receivedNVMC := lh.NewViewMessageContentReader(nvmBytes)
		require.Equal(t, receivedNVMC.SignedHeader().MessageType(), lh.LEAN_HELIX_NEW_VIEW, "Message type should be LEAN_HELIX_NEW_VIEW")
		require.True(t, receivedNVMC.SignedHeader().BlockHeight().Equal(height), "Height = %v", height)
		require.True(t, receivedNVMC.SignedHeader().View().Equal(view), "View = %v", view)
	})

	t.Run("create PreprepareMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppmcb := &lh.PreprepareMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: leaderKeyManager.MyPublicKey(),
				Signature:       leaderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		expectedPPM := &lh.PreprepareMessageImpl{
			Content: ppmcb.Build(),
			MyBlock: block,
		}

		actualPPM := leaderFac.CreatePreprepareMessage(height, view, block)
		expectedPPMRaw := expectedPPM.Raw()
		actualPPMRaw := actualPPM.Raw()

		require.True(t, bytes.Compare(expectedPPMRaw, actualPPMRaw) == 0, "compared bytes of PPM")
	})

	t.Run("create PrepareMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pmcb := &lh.PrepareMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: leaderKeyManager.MyPublicKey(),
				Signature:       leaderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}
		expectedPM := &lh.PrepareMessageImpl{
			Content: pmcb.Build(),
		}
		actualPM := leaderFac.CreatePrepareMessage(height, view, blockHash)
		expectedPMRaw := expectedPM.Raw()
		actualPMRaw := actualPM.Raw()
		require.True(t, bytes.Compare(expectedPMRaw, actualPMRaw) == 0, "compared bytes of PM")
	})
	t.Run("create CommitMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_COMMIT,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		cmcb := &lh.CommitMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: leaderKeyManager.MyPublicKey(),
				Signature:       leaderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}
		expectedCM := &lh.CommitMessageImpl{
			Content: cmcb.Build(),
		}
		actualCM := leaderFac.CreateCommitMessage(height, view, blockHash)
		expectedCMRaw := expectedCM.Raw()
		actualCMRaw := actualCM.Raw()
		require.True(t, bytes.Compare(expectedCMRaw, actualCMRaw) == 0, "compared bytes of CM")

	})

	// TODO This needs further testing - no proof, no pp or no p's with the proof
	t.Run("create ViewChangeMessage with PreparedProof", func(t *testing.T) {
		proofBuilder := lh.CreatePreparedProofBuilder(leaderKeyManager, []lh.KeyManager{node1KeyManager, node2KeyManager}, height, view, blockHash)

		signedHeader := &lh.ViewChangeHeaderBuilder{
			MessageType:   lh.LEAN_HELIX_VIEW_CHANGE,
			BlockHeight:   height,
			View:          view,
			PreparedProof: proofBuilder,
		}
		vcmcb := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: leaderKeyManager.MyPublicKey(),
				Signature:       leaderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}
		expectedVCM := &lh.ViewChangeMessageImpl{
			Content: vcmcb.Build(),
			MyBlock: block,
		}

		preparedMessages := &lh.PreparedMessages{
			PreprepareMessage: leaderFac.CreatePreprepareMessage(height, view, block),
			PrepareMessages: []lh.PrepareMessage{
				node1Fac.CreatePrepareMessage(height, view, blockHash),
				node2Fac.CreatePrepareMessage(height, view, blockHash),
			},
		}

		actualVCM := leaderFac.CreateViewChangeMessage(height, view, preparedMessages)
		fmt.Println(expectedVCM.String())
		fmt.Println(actualVCM.String())
		expectedVCMRaw := expectedVCM.Raw()
		actualVCMRaw := actualVCM.Raw()
		require.True(t, bytes.Compare(expectedVCMRaw, actualVCMRaw) == 0, "compared bytes of VCM")

	})

	t.Run("create ViewChangeMessage without PreparedProof", func(t *testing.T) {
		t.Skip()
	})

	t.Run("create NewViewMessage", func(t *testing.T) {

	})

}

// TODO VCM & NVM from MessageFactory.spec.ts
