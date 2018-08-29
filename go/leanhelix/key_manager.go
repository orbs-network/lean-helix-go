package leanhelix

type KeyManager interface {
	SignBlockMessageContent(bmc *BlockMessageContent) string
	SignViewChangeMessage(vcmc *ViewChangeMessageContent) string
	//SignNewViewMessage()
	MyPublicKey() PublicKey
	//
	//VerifyPrepreparePayload()
	//VerifyPreparePayload()
	//VerifyCommitPayload()
	//VerifyViewChangePayload()
	//VerifyNewViewPayload()
}
