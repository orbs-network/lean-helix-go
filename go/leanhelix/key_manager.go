package leanhelix

type KeyManager interface {
	SignBlockMessageContent(bmc *BlockMessageContent) string
	SignViewChangeMessage(vcmc *ViewChangeMessageContent) string

	VerifyBlockMessageContent(bmc *BlockMessageContent, signature string, publicKey PublicKey) bool

	MyPublicKey() PublicKey
}
