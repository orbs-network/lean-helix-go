package builders

//import "github.com/orbs-network/lean-helix-go"

//func CreatePreparedProofByMessages(preprepareMessage leanhelix.PreprepareMessage, prepareMessages []leanhelix.PrepareMessage) leanhelix.PreparedProof {
//
//	return leanhelix.PreparedProofBuilder{
//		PreprepareBlockRef: preprepareMessage,
//		PreprepareSender: &leanhelix.SenderSignatureBuilder{
//			SenderPublicKey: nil,
//			Signature:       nil,
//		},
//		PrepareBlockRef: &leanhelix.BlockRefBuilder{
//			MessageType: 0,
//			BlockHeight: 0,
//			View:        0,
//			BlockHash:   nil,
//		},
//		PrepareSenders: nil,
//	}
//
//	//	preprepareBlockRef: PPMessage ? PPMessage.signedHeader : undefined,
//	//preprepareSender: PPMessage ? PPMessage.sender : undefined,
//	//prepareBlockRef: PMessages ? PMessages[0].signedHeader : undefined,
//	//prepareSenders: PMessages ? PMessages.map(m => m.sender) : undefined
//	//};
//}
