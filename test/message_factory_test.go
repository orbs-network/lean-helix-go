package test

import (
	"bytes"
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	leaderKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK0"))
	node1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK1"))
	node2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK2"))
	height := BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := View(math.Floor(rand.Float64() * 1000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := block.BlockHash()
	leaderFac := lh.NewMessageFactory(leaderKeyManager)
	node1Fac := lh.NewMessageFactory(node1KeyManager)
	node2Fac := lh.NewMessageFactory(node2KeyManager)

	t.Run("create PreprepareMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppmcb := &lh.PreprepareContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: leaderKeyManager.MyPublicKey(),
				Signature:       leaderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		expectedPPM := lh.NewPreprepareMessage(ppmcb.Build(), block)

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
		prepareContentBuilder := &lh.PrepareContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: leaderKeyManager.MyPublicKey(),
				Signature:       leaderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}
		expectedPM := lh.NewPrepareMessage(prepareContentBuilder.Build())
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
		cmcb := &lh.CommitContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: leaderKeyManager.MyPublicKey(),
				Signature:       leaderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}
		expectedCM := lh.NewCommitMessage(cmcb.Build())
		actualCM := leaderFac.CreateCommitMessage(height, view, blockHash)
		expectedCMRaw := expectedCM.Raw()
		actualCMRaw := actualCM.Raw()
		require.True(t, bytes.Compare(expectedCMRaw, actualCMRaw) == 0, "compared bytes of CM")

	})

	// TODO This needs further testing - no proof, no pp or no p's with the proof
	t.Run("create ViewChangeMessage with PreparedProof", func(t *testing.T) {

		//proofBuilder := lh.CreatePreparedProofBuilder(leaderKeyManager, []lh.KeyManager{node1KeyManager, node2KeyManager}, height, view, blockHash)

		// Decide which node is sending this message
		senderKeyManager := node1KeyManager
		senderMessageFactory := node1Fac

		ppBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender := &lh.SenderSignatureBuilder{
			SenderPublicKey: leaderKeyManager.MyPublicKey(),
			Signature:       leaderKeyManager.Sign(ppBlockRefBuilder.Build().Raw()),
		}
		pBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders := []*lh.SenderSignatureBuilder{
			{
				SenderPublicKey: node1KeyManager.MyPublicKey(),
				Signature:       node1KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
			{
				SenderPublicKey: node2KeyManager.MyPublicKey(),
				Signature:       node2KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
		}
		proofBuilder := &lh.PreparedProofBuilder{
			PreprepareBlockRef: ppBlockRefBuilder,
			PreprepareSender:   ppSender,
			PrepareBlockRef:    pBlockRefBuilder,
			PrepareSenders:     pSenders,
		}
		signedHeader := &lh.ViewChangeHeaderBuilder{
			MessageType:   lh.LEAN_HELIX_VIEW_CHANGE,
			BlockHeight:   height,
			View:          view,
			PreparedProof: proofBuilder,
		}
		vcmContentBuilder := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: senderKeyManager.MyPublicKey(),
				Signature:       senderKeyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		expectedVCM := lh.NewViewChangeMessage(vcmContentBuilder.Build(), block)
		preparedMessages := &lh.PreparedMessages{
			PreprepareMessage: leaderFac.CreatePreprepareMessage(height, view, block),
			PrepareMessages: []*lh.PrepareMessage{
				node1Fac.CreatePrepareMessage(height, view, blockHash),
				node2Fac.CreatePrepareMessage(height, view, blockHash),
			},
		}
		actualVCM := senderMessageFactory.CreateViewChangeMessage(height, view, preparedMessages)
		fmt.Println("Expected:", expectedVCM.String())
		fmt.Println("Actual:  ", actualVCM.String())
		expectedVCMRaw := expectedVCM.Raw()
		actualVCMRaw := actualVCM.Raw()
		require.True(t, bytes.Compare(expectedVCMRaw, actualVCMRaw) == 0, "compared bytes of VCM")

	})

	t.Run("create ViewChangeMessage without PreparedProof", func(t *testing.T) {
		t.Skip()
	})

	t.Run("create NewViewMessage", func(t *testing.T) {

		// This test passes falsely!!!

		// This test assumes all non-leader nodes hold PreparedProof and also that it is the same for all of them

		// Construct the "expected" message manually
		ppBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender := &lh.SenderSignatureBuilder{
			SenderPublicKey: leaderKeyManager.MyPublicKey(),
			Signature:       leaderKeyManager.Sign(ppBlockRefBuilder.Build().Raw()),
		}
		pBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders := []*lh.SenderSignatureBuilder{
			{
				SenderPublicKey: node1KeyManager.MyPublicKey(),
				Signature:       node1KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
			{
				SenderPublicKey: node2KeyManager.MyPublicKey(),
				Signature:       node2KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
		}
		proofBuilder := &lh.PreparedProofBuilder{
			PreprepareBlockRef: ppBlockRefBuilder,
			PreprepareSender:   ppSender,
			PrepareBlockRef:    pBlockRefBuilder,
			PrepareSenders:     pSenders,
		}

		nodesVCHeader := &lh.ViewChangeHeaderBuilder{
			MessageType:   lh.LEAN_HELIX_VIEW_CHANGE,
			BlockHeight:   height,
			View:          view,
			PreparedProof: proofBuilder,
		}
		node1Confirmation := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: nodesVCHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: node1KeyManager.MyPublicKey(),
				Signature:       node1KeyManager.Sign(nodesVCHeader.Build().Raw()),
			},
		}
		node2Confirmation := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: nodesVCHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: node2KeyManager.MyPublicKey(),
				Signature:       node2KeyManager.Sign(nodesVCHeader.Build().Raw()),
			},
		}
		nvmHeader := &lh.NewViewHeaderBuilder{
			MessageType: lh.LEAN_HELIX_NEW_VIEW,
			BlockHeight: height,
			View:        view,
			ViewChangeConfirmations: []*lh.ViewChangeMessageContentBuilder{
				node1Confirmation, node2Confirmation,
			},
		}
		nvmSender := &lh.SenderSignatureBuilder{
			SenderPublicKey: leaderKeyManager.MyPublicKey(),
			Signature:       leaderKeyManager.Sign(nvmHeader.Build().Raw()),
		}
		nvmContentBuilder := &lh.NewViewMessageContentBuilder{
			SignedHeader: nvmHeader,
			Sender:       nvmSender,
			PreprepareMessageContent: &lh.PreprepareContentBuilder{
				SignedHeader: ppBlockRefBuilder,
				Sender:       ppSender,
			},
		}

		expectedNVM := lh.NewNewViewMessage(nvmContentBuilder.Build(), block)

		// Construct "actual" message with message factories
		ppm := leaderFac.CreatePreprepareMessage(height, view, block)
		preparedMessages := &lh.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages: []*lh.PrepareMessage{
				node1Fac.CreatePrepareMessage(height, view, blockHash),
				node2Fac.CreatePrepareMessage(height, view, blockHash),
			},
		}
		confirmations := []*lh.ViewChangeMessageContentBuilder{
			node1Fac.CreateViewChangeMessageContentBuilder(
				height, view, preparedMessages),
			node2Fac.CreateViewChangeMessageContentBuilder(
				height, view, preparedMessages),
		}
		actualNVM := leaderFac.CreateNewViewMessage(
			height,
			view,
			leaderFac.CreatePreprepareMessageContentBuilder(
				height, view, block),
			confirmations,
			block)

		fmt.Println(expectedNVM)
		fmt.Println(actualNVM)
		expectedNVMRaw := expectedNVM.Raw()
		actualNVMRaw := actualNVM.Raw()
		require.True(t, bytes.Compare(expectedNVMRaw, actualNVMRaw) == 0, "compared bytes of NVM")

	})

}

// TODO VCM & NVM from MessageFactory.spec.ts
