package messages_test

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/proto"
	"testing"
)

func CreatePreprepareMessage(
	utils leanhelix.BlockUtils,
	keyManager leanhelix.KeyManager,
	blockHeight leanhelix.BlockHeight,
	view leanhelix.View,
	block leanhelix.Block) *messages.LeanHelixPrePrepareMessage {

	var (
		header *messages.LeanHelixBlockRefBuilder
		sender *messages.LeanHelixSenderSignatureBuilder
	)

	header = &messages.LeanHelixBlockRefBuilder{
		MessageType: messages.LEAN_HELIX_PRE_PREPARE,
		BlockHeight: &messages.BlockHeightBuilder{
			Value: uint64(blockHeight),
		},
		View: uint32(view),
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

	return ppm.Build()
}

func TestCreatePPM(t *testing.T) {

	// TODO Write this thing

	//ppm := CreatePreprepareMessage(&builders.MockBlockUtils{}, )

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
