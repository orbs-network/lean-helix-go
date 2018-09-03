package keymanagermock

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go/go/leanhelix"
)

// TODO Keys should not be strings - convert to our primitives

const PRIVATE_KEY_PREFIX = "PRIVATE_KEY"

type mockKeyManager struct {
	mock.Mock
	myPublicKey        leanhelix.PublicKey
	RejectedPublicKeys []leanhelix.PublicKey
}

func NewMockKeyManager(publicKey leanhelix.PublicKey, rejectedPublicKeys ...leanhelix.PublicKey) *mockKeyManager {
	return &mockKeyManager{
		myPublicKey:        publicKey,
		RejectedPublicKeys: rejectedPublicKeys,
	}
}

func (km *mockKeyManager) SignViewChangeMessage(vcmc *leanhelix.ViewChangeMessageContent) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", vcmc.MessageType, PRIVATE_KEY_PREFIX, km.MyPublicKey(), string(vcmc.Term), string(vcmc.View))
}

func (km *mockKeyManager) SignBlockMessageContent(bmc *leanhelix.BlockMessageContent) string {
	return fmt.Sprintf("%s|%s|%s|%d|%d|%s", bmc.MessageType, PRIVATE_KEY_PREFIX, km.MyPublicKey(), bmc.Term, bmc.View, bmc.BlockHash)
}

func (km *mockKeyManager) VerifyBlockMessageContent(bmc *leanhelix.BlockMessageContent, signature string, publicKey leanhelix.PublicKey) bool {
	for _, rejectedKey := range km.RejectedPublicKeys {
		if rejectedKey == publicKey {
			return false
		}
	}

	signedMessage := fmt.Sprintf("%s|%s|%s|%d|%d|%s", bmc.MessageType, PRIVATE_KEY_PREFIX, publicKey, bmc.Term, bmc.View, bmc.BlockHash)
	return signedMessage == signature
}

func (km *mockKeyManager) MyPublicKey() leanhelix.PublicKey {
	return km.myPublicKey
}
