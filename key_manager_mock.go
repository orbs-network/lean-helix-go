package leanhelix

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go/types"
)

// TODO Keys should not be strings - convert to our primitives

const PRIVATE_KEY_PREFIX = "PRIVATE_KEY"

type mockKeyManager struct {
	mock.Mock
	myPublicKey        types.PublicKey
	RejectedPublicKeys []types.PublicKey
}

func NewMockKeyManager(publicKey types.PublicKey, rejectedPublicKeys ...types.PublicKey) *mockKeyManager {
	return &mockKeyManager{
		myPublicKey:        publicKey,
		RejectedPublicKeys: rejectedPublicKeys,
	}
}

func (km *mockKeyManager) SignViewChangeMessage(vcmc *ViewChangeMessageContent) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", vcmc.MessageType, PRIVATE_KEY_PREFIX, km.MyPublicKey(), string(vcmc.Term), string(vcmc.View))
}

func (km *mockKeyManager) SignBlockMessageContent(bmc *BlockMessageContent) string {
	return fmt.Sprintf("%s|%s|%s|%d|%d|%s", bmc.MessageType, PRIVATE_KEY_PREFIX, km.MyPublicKey(), bmc.Term, bmc.View, bmc.BlockHash)
}

func (km *mockKeyManager) VerifyBlockMessageContent(bmc *BlockMessageContent, signature string, publicKey types.PublicKey) bool {
	for _, rejectedKey := range km.RejectedPublicKeys {
		if rejectedKey == publicKey {
			return false
		}
	}

	signedMessage := fmt.Sprintf("%s|%s|%s|%d|%d|%s", bmc.MessageType, PRIVATE_KEY_PREFIX, publicKey, bmc.Term, bmc.View, bmc.BlockHash)
	return signedMessage == signature
}

func (km *mockKeyManager) MyPublicKey() types.PublicKey {
	return km.myPublicKey
}
