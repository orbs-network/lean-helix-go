package test

import (
	"bytes"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	memberId := primitives.MemberId("Member Id0")
	memberId1 := primitives.MemberId("Member Id1")
	memberId2 := primitives.MemberId("Member Id2")

	keyManager := builders.NewMockKeyManager(memberId)
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	node1KeyManager := builders.NewMockKeyManager(memberId1)
	node2KeyManager := builders.NewMockKeyManager(memberId2)
	leaderFac := leanhelix.NewMessageFactory(keyManager, memberId)
	node1Factory := leanhelix.NewMessageFactory(node1KeyManager, memberId1)
	node2Fac := leanhelix.NewMessageFactory(node2KeyManager, memberId2)

	t.Run("create PreprepareMessage", func(t *testing.T) {
		signedHeader := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		ppmcb := &protocol.PreprepareContentBuilder{
			SignedHeader: signedHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId,
				Signature: keyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		expectedPPM := leanhelix.NewPreprepareMessage(ppmcb.Build(), block)
		actualPPM := leaderFac.CreatePreprepareMessage(blockHeight, view, block, blockHash)

		require.True(t, bytes.Compare(expectedPPM.Raw(), actualPPM.Raw()) == 0, "compared bytes of PPM")
	})

	t.Run("create PrepareMessage", func(t *testing.T) {
		signedHeader := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		prepareContentBuilder := &protocol.PrepareContentBuilder{
			SignedHeader: signedHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId,
				Signature: keyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		expectedPM := leanhelix.NewPrepareMessage(prepareContentBuilder.Build())
		actualPM := leaderFac.CreatePrepareMessage(blockHeight, view, blockHash)

		require.True(t, bytes.Compare(expectedPM.Raw(), actualPM.Raw()) == 0, "compared bytes of PM")
	})

	t.Run("create CommitMessage", func(t *testing.T) {
		signedHeader := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_COMMIT,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		cmcb := &protocol.CommitContentBuilder{
			SignedHeader: signedHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId,
				Signature: keyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		expectedCM := leanhelix.NewCommitMessage(cmcb.Build())
		actualCM := leaderFac.CreateCommitMessage(blockHeight, view, blockHash)

		require.True(t, bytes.Compare(expectedCM.Raw(), actualCM.Raw()) == 0, "compared bytes of CM")

	})

	t.Run("create ViewChangeMessage without PreparedProof", func(t *testing.T) {
		signedHeader := &protocol.ViewChangeHeaderBuilder{
			MessageType:   protocol.LEAN_HELIX_VIEW_CHANGE,
			BlockHeight:   blockHeight,
			View:          view,
			PreparedProof: nil,
		}
		vcmContentBuilder := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId1,
				Signature: node1KeyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		actualVCM := node1Factory.CreateViewChangeMessage(blockHeight, view, nil)
		expectedVCM := leanhelix.NewViewChangeMessage(vcmContentBuilder.Build(), nil)

		require.True(t, bytes.Compare(expectedVCM.Raw(), actualVCM.Raw()) == 0, "compared bytes of VCM")
	})

	t.Run("create ViewChangeMessage with PreparedProof", func(t *testing.T) {
		ppBlockRefBuilder := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender := &protocol.SenderSignatureBuilder{
			MemberId:  memberId,
			Signature: keyManager.Sign(ppBlockRefBuilder.Build().Raw()),
		}
		pBlockRefBuilder := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders := []*protocol.SenderSignatureBuilder{
			{
				MemberId:  memberId1,
				Signature: node1KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
			{
				MemberId:  memberId2,
				Signature: node2KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
		}
		proofBuilder := &protocol.PreparedProofBuilder{
			PreprepareBlockRef: ppBlockRefBuilder,
			PreprepareSender:   ppSender,
			PrepareBlockRef:    pBlockRefBuilder,
			PrepareSenders:     pSenders,
		}
		signedHeader := &protocol.ViewChangeHeaderBuilder{
			MessageType:   protocol.LEAN_HELIX_VIEW_CHANGE,
			BlockHeight:   blockHeight,
			View:          view,
			PreparedProof: proofBuilder,
		}
		vcmContentBuilder := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId1,
				Signature: node1KeyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		preparedMessages := &leanhelix.PreparedMessages{
			PreprepareMessage: leaderFac.CreatePreprepareMessage(blockHeight, view, block, blockHash),
			PrepareMessages: []*leanhelix.PrepareMessage{
				node1Factory.CreatePrepareMessage(blockHeight, view, blockHash),
				node2Fac.CreatePrepareMessage(blockHeight, view, blockHash),
			},
		}

		actualVCM := node1Factory.CreateViewChangeMessage(blockHeight, view, preparedMessages)
		expectedVCM := leanhelix.NewViewChangeMessage(vcmContentBuilder.Build(), block)

		require.True(t, bytes.Compare(expectedVCM.Raw(), actualVCM.Raw()) == 0, "compared bytes of VCM")
	})

	t.Run("create NewViewMessage", func(t *testing.T) {
		// This test assumes all non-leader nodes hold PreparedProof and also that it is the same for all of them

		// Construct the "expected" message manually
		ppBlockRefBuilder := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		ppSender := &protocol.SenderSignatureBuilder{
			MemberId:  memberId,
			Signature: keyManager.Sign(ppBlockRefBuilder.Build().Raw()),
		}
		pBlockRefBuilder := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_PREPARE,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}
		pSenders := []*protocol.SenderSignatureBuilder{
			{
				MemberId:  memberId1,
				Signature: node1KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
			{
				MemberId:  memberId2,
				Signature: node2KeyManager.Sign(pBlockRefBuilder.Build().Raw()),
			},
		}
		proofBuilder := &protocol.PreparedProofBuilder{
			PreprepareBlockRef: ppBlockRefBuilder,
			PreprepareSender:   ppSender,
			PrepareBlockRef:    pBlockRefBuilder,
			PrepareSenders:     pSenders,
		}

		nodesVCHeader := &protocol.ViewChangeHeaderBuilder{
			MessageType:   protocol.LEAN_HELIX_VIEW_CHANGE,
			BlockHeight:   blockHeight,
			View:          view,
			PreparedProof: proofBuilder,
		}
		node1Confirmation := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: nodesVCHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId1,
				Signature: node1KeyManager.Sign(nodesVCHeader.Build().Raw()),
			},
		}
		node2Confirmation := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: nodesVCHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId2,
				Signature: node2KeyManager.Sign(nodesVCHeader.Build().Raw()),
			},
		}
		nvmHeader := &protocol.NewViewHeaderBuilder{
			MessageType: protocol.LEAN_HELIX_NEW_VIEW,
			BlockHeight: blockHeight,
			View:        view,
			ViewChangeConfirmations: []*protocol.ViewChangeMessageContentBuilder{
				node1Confirmation, node2Confirmation,
			},
		}
		nvmSender := &protocol.SenderSignatureBuilder{
			MemberId:  memberId,
			Signature: keyManager.Sign(nvmHeader.Build().Raw()),
		}
		nvmContentBuilder := &protocol.NewViewMessageContentBuilder{
			SignedHeader: nvmHeader,
			Sender:       nvmSender,
			Message: &protocol.PreprepareContentBuilder{
				SignedHeader: ppBlockRefBuilder,
				Sender:       ppSender,
			},
		}

		// Construct "actual" message with message factories
		ppm := leaderFac.CreatePreprepareMessage(blockHeight, view, block, blockHash)
		preparedMessages := &leanhelix.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages: []*leanhelix.PrepareMessage{
				node1Factory.CreatePrepareMessage(blockHeight, view, blockHash),
				node2Fac.CreatePrepareMessage(blockHeight, view, blockHash),
			},
		}
		confirmations := []*protocol.ViewChangeMessageContentBuilder{
			node1Factory.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages),
			node2Fac.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages),
		}

		actualNVM := leaderFac.CreateNewViewMessage(
			blockHeight,
			view,
			leaderFac.CreatePreprepareMessageContentBuilder(blockHeight, view, block, blockHash),
			confirmations,
			block)
		expectedNVM := leanhelix.NewNewViewMessage(nvmContentBuilder.Build(), block)

		require.True(t, bytes.Compare(expectedNVM.Raw(), actualNVM.Raw()) == 0, "compared bytes of NVM")

	})

}
