package leanhelix_test

import (
	"github.com/orbs-network/lean-helix-go"
	"testing"
)

func CreatePreprepareMessage(
	utils leanhelix.BlockUtils,
	keyManager leanhelix.KeyManager,
	blockHeight uint64,
	view uint64,
	block leanhelix.Block) *messages.LeanHelixPrePrepareMessageBuilder {

	var (
		header *messages.LeanHelixBlockRefBuilder
		sender *messages.LeanHelixSenderSignatureBuilder
	)

	header = &messages.LeanHelixBlockRefBuilder{
		MessageType: messages.LEAN_HELIX_PRE_PREPARE,
		BlockHeight: &messages.BlockHeightBuilder{
			Value: uint64(blockHeight),
		},
		View: &messages.ViewBuilder{
			Value: uint64(view),
		},
		BlockHash: &messages.Uint256Builder{
			Value: utils.CalculateBlockHash(block),
		},
	}

	sig := &messages.Ed25519_sigBuilder{Value: keyManager.Sign(header.Build().Raw())}
	me := &messages.Ed25519_public_keyBuilder{Value: keyManager.MyPublicKey()}
	sender = &messages.LeanHelixSenderSignatureBuilder{
		SenderPublicKey: me,
		Signature:       sig,
	}

	ppm := &messages.LeanHelixPrePrepareMessageBuilder{
		SignedHeader: header,
		Sender:       sender,
	}

	return ppm
}

func TestCreatePPM(t *testing.T) {

	// TODO This will demo how to create a message and convert to []byte
	// TODO Uncomment and write it

	//ppm := CreatePreprepareMessage(&builders.MockBlockUtils{}, )
	//bytes := ppm.Raw()
	//
}

//func CreatePreparedProof(ref leanhelix.BlockRef) {
//	res := (messages.LeanHelixPreparedProofBuilder{
//		BlockRef: &messages.LeanHelixBlockRefBuilder{
//			MessageType: ref.MessageType(),
//			BlockHeight: 0,
//			View:        0,
//			BlockHash:   nil,
//		},
//		Senders: nil,
//	}).Build()
//	return (*PreparedProofImpl)(res)
//}
//
