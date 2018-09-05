package leanhelix

import "github.com/orbs-network/lean-helix-go/types"

type KeyManager interface {
	SignBlockMessageContent(bmc *BlockMessageContent) string
	SignViewChangeMessage(vcmc *ViewChangeMessageContent) string

	VerifyBlockMessageContent(bmc *BlockMessageContent, signature string, publicKey types.PublicKey) bool

	MyPublicKey() types.PublicKey
}
