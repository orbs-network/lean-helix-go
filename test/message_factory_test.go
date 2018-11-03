package test

import (
	"bytes"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	keyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK0"))
	blockHeight := BlockHeight(math.Floor(rand.Float64() * 1000000000))
	view := View(math.Floor(rand.Float64() * 1000000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := block.BlockHash()
	node1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK1"))
	node2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK2"))
	leaderFac := lh.NewMessageFactory(keyManager)
	node1Fac := lh.NewMessageFactory(node1KeyManager)
	node2Fac := lh.NewMessageFactory(node2KeyManager)

	t.Run("create PreprepareMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		dataToSign := signedHeader.Build().Raw()
		sig, err := keyManager.Sign(dataToSign)
		if err != nil {
			t.Error(err)
		}
		ppmcb := &lh.PreprepareContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: keyManager.MyPublicKey(),
				Signature:       sig,
			},
		}

		expectedPPM := lh.NewPreprepareMessage(ppmcb.Build(), block)
		actualPPM := leaderFac.CreatePreprepareMessage(blockHeight, view, block)

		require.True(t, bytes.Compare(expectedPPM.Raw(), actualPPM.Raw()) == 0, "compared bytes of PPM")
	})

	t.Run("create PrepareMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		dataToSign := signedHeader.Build().Raw()
		sig, err := keyManager.Sign(dataToSign)
		if err != nil {
			t.Error(err)
		}
		prepareContentBuilder := &lh.PrepareContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: keyManager.MyPublicKey(),
				Signature:       sig,
			},
		}

		expectedPM := lh.NewPrepareMessage(prepareContentBuilder.Build())
		actualPM := leaderFac.CreatePrepareMessage(blockHeight, view, blockHash)

		require.True(t, bytes.Compare(expectedPM.Raw(), actualPM.Raw()) == 0, "compared bytes of PM")
	})

	t.Run("create CommitMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_COMMIT,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		dataToSign := signedHeader.Build().Raw()
		sig, err := keyManager.Sign(dataToSign)
		if err != nil {
			t.Error(err)
		}
		cmcb := &lh.CommitContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: keyManager.MyPublicKey(),
				Signature:       sig,
			},
		}

		expectedCM := lh.NewCommitMessage(cmcb.Build())
		actualCM := leaderFac.CreateCommitMessage(blockHeight, view, blockHash)

		require.True(t, bytes.Compare(expectedCM.Raw(), actualCM.Raw()) == 0, "compared bytes of CM")

	})

	t.Run("create ViewChangeMessage without PreparedProof", func(t *testing.T) {
		senderKeyManager := node1KeyManager
		senderMessageFactory := node1Fac

		signedHeader := &lh.ViewChangeHeaderBuilder{
			MessageType:   lh.LEAN_HELIX_VIEW_CHANGE,
			BlockHeight:   blockHeight,
			View:          view,
			PreparedProof: nil,
		}
		dataToSign := signedHeader.Build().Raw()
		sig, err := senderKeyManager.Sign(dataToSign)
		if err != nil {
			t.Error(err)
		}
		vcmContentBuilder := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: senderKeyManager.MyPublicKey(),
				Signature:       sig,
			},
		}

		actualVCM := senderMessageFactory.CreateViewChangeMessage(blockHeight, view, nil)
		expectedVCM := lh.NewViewChangeMessage(vcmContentBuilder.Build(), nil)

		require.True(t, bytes.Compare(expectedVCM.Raw(), actualVCM.Raw()) == 0, "compared bytes of VCM")
	})

	t.Run("create ViewChangeMessage with PreparedProof", func(t *testing.T) {
		senderKeyManager := node1KeyManager
		senderMessageFactory := node1Fac

		ppBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		ppDataToSign := ppBlockRefBuilder.Build().Raw()
		sig, err := keyManager.Sign(ppDataToSign)
		if err != nil {
			t.Error(err)
		}
		ppSender := &lh.SenderSignatureBuilder{
			SenderPublicKey: keyManager.MyPublicKey(),
			Signature:       sig,
		}
		pBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}

		pDataToSign := pBlockRefBuilder.Build().Raw()
		sig1, err := node1KeyManager.Sign(pDataToSign)
		if err != nil {
			t.Error(err)
		}
		sig2, err := node2KeyManager.Sign(pDataToSign)
		if err != nil {
			t.Error(err)
		}
		pSenders := []*lh.SenderSignatureBuilder{
			{
				SenderPublicKey: node1KeyManager.MyPublicKey(),
				Signature:       sig1,
			},
			{
				SenderPublicKey: node2KeyManager.MyPublicKey(),
				Signature:       sig2,
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
			BlockHeight:   blockHeight,
			View:          view,
			PreparedProof: proofBuilder,
		}

		vcmDataToSign := signedHeader.Build().Raw()
		vcmSig, err := node1KeyManager.Sign(vcmDataToSign)
		if err != nil {
			t.Error(err)
		}
		vcmContentBuilder := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: senderKeyManager.MyPublicKey(),
				Signature:       vcmSig,
			},
		}

		preparedMessages := &lh.PreparedMessages{
			PreprepareMessage: leaderFac.CreatePreprepareMessage(blockHeight, view, block),
			PrepareMessages: []*lh.PrepareMessage{
				node1Fac.CreatePrepareMessage(blockHeight, view, blockHash),
				node2Fac.CreatePrepareMessage(blockHeight, view, blockHash),
			},
		}

		actualVCM := senderMessageFactory.CreateViewChangeMessage(blockHeight, view, preparedMessages)
		expectedVCM := lh.NewViewChangeMessage(vcmContentBuilder.Build(), block)

		require.True(t, bytes.Compare(expectedVCM.Raw(), actualVCM.Raw()) == 0, "compared bytes of VCM")
	})

	t.Run("create NewViewMessage", func(t *testing.T) {
		// This test assumes all non-leader nodes hold PreparedProof and also that it is the same for all of them

		// Construct the "expected" message manually
		ppBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}

		ppDataToSign := ppBlockRefBuilder.Build().Raw()
		sig, err := keyManager.Sign(ppDataToSign)
		if err != nil {
			t.Error(err)
		}
		ppSender := &lh.SenderSignatureBuilder{
			SenderPublicKey: keyManager.MyPublicKey(),
			Signature:       sig,
		}
		pBlockRefBuilder := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		pDataToSign := pBlockRefBuilder.Build().Raw()
		sig1, err := node1KeyManager.Sign(pDataToSign)
		if err != nil {
			t.Error(err)
		}
		sig2, err := node2KeyManager.Sign(pDataToSign)
		if err != nil {
			t.Error(err)
		}
		pSenders := []*lh.SenderSignatureBuilder{
			{
				SenderPublicKey: node1KeyManager.MyPublicKey(),
				Signature:       sig1,
			},
			{
				SenderPublicKey: node2KeyManager.MyPublicKey(),
				Signature:       sig2,
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
			BlockHeight:   blockHeight,
			View:          view,
			PreparedProof: proofBuilder,
		}
		vcDataToSign := nodesVCHeader.Build().Raw()
		vcSig1, err := node1KeyManager.Sign(vcDataToSign)
		if err != nil {
			t.Error(err)
		}
		vcSig2, err := node2KeyManager.Sign(vcDataToSign)
		if err != nil {
			t.Error(err)
		}
		node1Confirmation := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: nodesVCHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: node1KeyManager.MyPublicKey(),
				Signature:       vcSig1,
			},
		}
		node2Confirmation := &lh.ViewChangeMessageContentBuilder{
			SignedHeader: nodesVCHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: node2KeyManager.MyPublicKey(),
				Signature:       vcSig2,
			},
		}
		nvmHeader := &lh.NewViewHeaderBuilder{
			MessageType: lh.LEAN_HELIX_NEW_VIEW,
			BlockHeight: blockHeight,
			View:        view,
			ViewChangeConfirmations: []*lh.ViewChangeMessageContentBuilder{
				node1Confirmation, node2Confirmation,
			},
		}
		nvmDataToSign := nvmHeader.Build().Raw()
		nvSig, err := keyManager.Sign(nvmDataToSign)
		if err != nil {
			t.Error(err)
		}
		nvmSender := &lh.SenderSignatureBuilder{
			SenderPublicKey: keyManager.MyPublicKey(),
			Signature:       nvSig,
		}
		nvmContentBuilder := &lh.NewViewMessageContentBuilder{
			SignedHeader: nvmHeader,
			Sender:       nvmSender,
			PreprepareMessageContent: &lh.PreprepareContentBuilder{
				SignedHeader: ppBlockRefBuilder,
				Sender:       ppSender,
			},
		}

		// Construct "actual" message with message factories
		ppm := leaderFac.CreatePreprepareMessage(blockHeight, view, block)
		preparedMessages := &lh.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages: []*lh.PrepareMessage{
				node1Fac.CreatePrepareMessage(blockHeight, view, blockHash),
				node2Fac.CreatePrepareMessage(blockHeight, view, blockHash),
			},
		}
		confirmations := []*lh.ViewChangeMessageContentBuilder{
			node1Fac.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages),
			node2Fac.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages),
		}

		actualNVM := leaderFac.CreateNewViewMessage(
			blockHeight,
			view,
			leaderFac.CreatePreprepareMessageContentBuilder(blockHeight, view, block),
			confirmations,
			block)
		expectedNVM := lh.NewNewViewMessage(nvmContentBuilder.Build(), block)

		require.True(t, bytes.Compare(expectedNVM.Raw(), actualNVM.Raw()) == 0, "compared bytes of NVM")

	})

}
