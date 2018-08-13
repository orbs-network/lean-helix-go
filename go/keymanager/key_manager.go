package keymanager

type KeyManager interface {
	SignPrepreparePayload()
	VerifyPrepreparePayload()

	SignPreparePayload()
	VerifyPreparePayload()

	SignCommitPayload()
	VerifyCommitPayload()

	SignViewChangePayload()
	VerifyViewChangePayload()

	SignNewViewPayload()
	VerifyNewViewPayload()
}
