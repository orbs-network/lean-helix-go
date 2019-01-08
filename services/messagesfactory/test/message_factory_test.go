package test

import (
	"bytes"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	networkId := primitives.NetworkId(rand.Uint64())
	memberId0 := primitives.MemberId("Member Id0")
	memberId1 := primitives.MemberId("Member Id1")
	memberId2 := primitives.MemberId("Member Id2")

	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	block := mocks.ABlock(interfaces.GenesisBlock)
	blockHash := mocks.CalculateBlockHash(block)

	node0KeyManager := mocks.NewMockKeyManager(memberId0)
	node1KeyManager := mocks.NewMockKeyManager(memberId1)
	node2KeyManager := mocks.NewMockKeyManager(memberId2)

	randomSeed := uint64(678)
	node0Factory := messagesfactory.NewMessageFactory(networkId, node0KeyManager, memberId0, randomSeed)
	node1Factory := messagesfactory.NewMessageFactory(networkId, node1KeyManager, memberId1, randomSeed)
	node2Factory := messagesfactory.NewMessageFactory(networkId, node2KeyManager, memberId2, randomSeed)

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
				MemberId:  memberId0,
				Signature: node0KeyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw()),
			},
		}

		expectedPPM := interfaces.NewPreprepareMessage(ppmcb.Build(), block)
		actualPPM := node0Factory.CreatePreprepareMessage(blockHeight, view, block, blockHash)

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
				MemberId:  memberId0,
				Signature: node0KeyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw()),
			},
		}

		expectedPM := interfaces.NewPrepareMessage(prepareContentBuilder.Build())
		actualPM := node0Factory.CreatePrepareMessage(blockHeight, view, blockHash)

		require.True(t, bytes.Compare(expectedPM.Raw(), actualPM.Raw()) == 0, "compared bytes of PM")
	})

	t.Run("create CommitMessage", func(t *testing.T) {
		signedHeader := &protocol.BlockRefBuilder{
			MessageType: protocol.LEAN_HELIX_COMMIT,
			BlockHeight: blockHeight,
			View:        view,
			BlockHash:   blockHash,
		}

		randomSeedBytes := randomseed.RandomSeedToBytes(randomSeed)
		share := node0KeyManager.SignRandomSeed(blockHeight, randomSeedBytes)
		cmcb := &protocol.CommitContentBuilder{
			SignedHeader: signedHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId0,
				Signature: node0KeyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw()),
			},
			Share: share,
		}

		expectedCM := interfaces.NewCommitMessage(cmcb.Build())
		actualCM := node0Factory.CreateCommitMessage(blockHeight, view, blockHash)

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
				Signature: node1KeyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw()),
			},
		}

		actualVCM := node1Factory.CreateViewChangeMessage(blockHeight, view, nil)
		expectedVCM := interfaces.NewViewChangeMessage(vcmContentBuilder.Build(), nil)

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
			MemberId:  memberId0,
			Signature: node0KeyManager.SignConsensusMessage(blockHeight, ppBlockRefBuilder.Build().Raw()),
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
				Signature: node1KeyManager.SignConsensusMessage(blockHeight, pBlockRefBuilder.Build().Raw()),
			},
			{
				MemberId:  memberId2,
				Signature: node2KeyManager.SignConsensusMessage(blockHeight, pBlockRefBuilder.Build().Raw()),
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
				Signature: node1KeyManager.SignConsensusMessage(blockHeight, signedHeader.Build().Raw()),
			},
		}

		preparedMessages := &preparedmessages.PreparedMessages{
			PreprepareMessage: node0Factory.CreatePreprepareMessage(blockHeight, view, block, blockHash),
			PrepareMessages: []*interfaces.PrepareMessage{
				node1Factory.CreatePrepareMessage(blockHeight, view, blockHash),
				node2Factory.CreatePrepareMessage(blockHeight, view, blockHash),
			},
		}

		actualVCM := node1Factory.CreateViewChangeMessage(blockHeight, view, preparedMessages)
		expectedVCM := interfaces.NewViewChangeMessage(vcmContentBuilder.Build(), block)

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
			MemberId:  memberId0,
			Signature: node0KeyManager.SignConsensusMessage(blockHeight, ppBlockRefBuilder.Build().Raw()),
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
				Signature: node1KeyManager.SignConsensusMessage(blockHeight, pBlockRefBuilder.Build().Raw()),
			},
			{
				MemberId:  memberId2,
				Signature: node2KeyManager.SignConsensusMessage(blockHeight, pBlockRefBuilder.Build().Raw()),
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
				Signature: node1KeyManager.SignConsensusMessage(blockHeight, nodesVCHeader.Build().Raw()),
			},
		}
		node2Confirmation := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: nodesVCHeader,
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  memberId2,
				Signature: node2KeyManager.SignConsensusMessage(blockHeight, nodesVCHeader.Build().Raw()),
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
			MemberId:  memberId0,
			Signature: node0KeyManager.SignConsensusMessage(blockHeight, nvmHeader.Build().Raw()),
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
		ppm := node0Factory.CreatePreprepareMessage(blockHeight, view, block, blockHash)
		preparedMessages := &preparedmessages.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages: []*interfaces.PrepareMessage{
				node1Factory.CreatePrepareMessage(blockHeight, view, blockHash),
				node2Factory.CreatePrepareMessage(blockHeight, view, blockHash),
			},
		}
		confirmations := []*protocol.ViewChangeMessageContentBuilder{
			node1Factory.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages),
			node2Factory.CreateViewChangeMessageContentBuilder(blockHeight, view, preparedMessages),
		}

		actualNVM := node0Factory.CreateNewViewMessage(
			blockHeight,
			view,
			node0Factory.CreatePreprepareMessageContentBuilder(blockHeight, view, block, blockHash),
			confirmations,
			block)
		expectedNVM := interfaces.NewNewViewMessage(nvmContentBuilder.Build(), block)

		require.True(t, bytes.Compare(expectedNVM.Raw(), actualNVM.Raw()) == 0, "compared bytes of NVM")

	})

}
