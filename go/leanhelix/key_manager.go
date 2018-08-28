package leanhelix

type KeyManager interface {
	SignBlockMessageContent(bmc *BlockMessageContent) string
	//SignViewChangeMessage()
	//SignNewViewMessage()
	MyPublicKey() PublicKey
	//
	//VerifyPrepreparePayload()
	//VerifyPreparePayload()
	//VerifyCommitPayload()
	//VerifyViewChangePayload()
	//VerifyNewViewPayload()
}
