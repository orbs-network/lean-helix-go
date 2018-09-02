package leanhelix

type KeyManager interface {
	SignBlockMessageContent(bmc *BlockMessageContent) string
	SignViewChangeMessage(vcmc *ViewChangeMessageContent) string
	//SignNewViewMessage()
	MyPublicKey() PublicKey

	VerifyBlockMessageContent(bmc *BlockMessageContent, signature string, publicKey PublicKey) bool
	//VerifyPrepreparePayload()
	//VerifyPreparePayload()
	//VerifyCommitPayload()
	//VerifyViewChangePayload()
	//VerifyNewViewPayload()
}
