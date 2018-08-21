package leanhelix

type KeyManager interface {
	SignPrepreparePayloadData(ppd *PrepreparePayloadData) string
	SignPreparePayloadData(pd *PreparePayloadData) string
	SignCommitPayloadData(cd *CommitPayloadData) string
	MyPublicKey() []byte
	//SignViewChangePayload()
	//SignNewViewPayload()
	//
	//VerifyPrepreparePayload()
	//VerifyPreparePayload()
	//VerifyCommitPayload()
	//VerifyViewChangePayload()
	//VerifyNewViewPayload()
}
