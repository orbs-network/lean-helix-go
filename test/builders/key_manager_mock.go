package builders

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
)

// TODO Keys should not be strings - convert to our primitives

const PRIVATE_KEY_PREFIX = "PRIVATE_KEY"

type mockKeyManager struct {
	mock.Mock
	myPublicKey        lh.PublicKey
	RejectedPublicKeys []lh.PublicKey
}

func (km *mockKeyManager) MyID() lh.PublicKey {
	return km.myPublicKey
}

func NewMockKeyManager(publicKey lh.PublicKey, rejectedPublicKeys ...lh.PublicKey) *mockKeyManager {
	return &mockKeyManager{
		myPublicKey:        publicKey,
		RejectedPublicKeys: rejectedPublicKeys,
	}
}

func (km *mockKeyManager) SignBlockRef(blockRef lh.BlockRef) lh.SenderSignature {
	return NewMockSenderSignature(km.MyID(),
		lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d|%s", blockRef.MessageType(), PRIVATE_KEY_PREFIX, km.MyID(), blockRef.Term(), blockRef.View(), blockRef.BlockHash())))
}

func (km *mockKeyManager) SignViewChange(vcm lh.ViewChangeMessage) lh.SenderSignature {
	return NewMockSenderSignature(km.MyID(),
		lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d", vcm.MessageType(), PRIVATE_KEY_PREFIX, km.MyID(), vcm.Term(), vcm.View())))
}

func (km *mockKeyManager) SignNewView(nvm lh.NewViewMessage) lh.SenderSignature {
	return NewMockSenderSignature(km.MyID(),
		lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d", nvm.MessageType(), PRIVATE_KEY_PREFIX, km.MyID(), nvm.Term(), nvm.View())))
}

func (km *mockKeyManager) VerifyBlockRef(blockRef lh.BlockRef, sender lh.SenderSignature) bool {

	if myIdRejected(sender.SenderPublicKey(), km.RejectedPublicKeys) {
		return false
	}

	signedMessage := lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d|%s", blockRef.MessageType(), PRIVATE_KEY_PREFIX, sender.SenderPublicKey(), blockRef.Term(), blockRef.View(), blockRef.BlockHash()))
	return signedMessage.Equals(sender.Signature())
}

func (km *mockKeyManager) VerifyViewChange(vcm lh.ViewChangeMessage, sender lh.SenderSignature) bool {
	if myIdRejected(sender.SenderPublicKey(), km.RejectedPublicKeys) {
		return false
	}

	signedMessage := lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d|%s", vcm.MessageType(), PRIVATE_KEY_PREFIX, sender.SenderPublicKey(), vcm.Term(), vcm.View(), vcm.BlockHash()))
	return signedMessage.Equals(sender.Signature())
}

func (km *mockKeyManager) VerifyNewView(nvm lh.NewViewMessage, sender lh.SenderSignature) bool {
	if myIdRejected(sender.SenderPublicKey(), km.RejectedPublicKeys) {
		return false
	}

	signedMessage := lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d|%s", nvm.MessageType(), PRIVATE_KEY_PREFIX, sender.SenderPublicKey(), nvm.Term(), nvm.View(), nvm.BlockHash()))
	return signedMessage.Equals(sender.Signature())
}

func myIdRejected(id lh.PublicKey, rejected []lh.PublicKey) bool {
	for _, rejectedKey := range rejected {
		if rejectedKey.Equals(id) {
			return true
		}
	}
	return false
}
