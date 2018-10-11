package builders

import (
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

func NewMockKeyManager(publicKey lh.PublicKey, rejectedPublicKeys ...lh.PublicKey) *mockKeyManager {
	return &mockKeyManager{
		myPublicKey:        publicKey,
		RejectedPublicKeys: rejectedPublicKeys,
	}
}

var MOCK_SIG_PREFIX = []byte("SIG|")

func (km *mockKeyManager) Sign(content []byte) []byte {
	return append(MOCK_SIG_PREFIX, content...)
}

func (km *mockKeyManager) Verify(content []byte, sender lh.SenderSignature) bool {
	panic("implement me")
}

func (km *mockKeyManager) MyPublicKey() lh.PublicKey {
	return km.myPublicKey
}

//func (km *mockKeyManager) SignBlockRef(blockRef lh.BlockRef) lh.SenderSignature {
//	return NewMockSenderSignature(km.MyPublicKey(),
//		lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d|%s", blockRef.MessageType(), PRIVATE_KEY_PREFIX, km.MyPublicKey(), blockRef.BlockHeight(), blockRef.View(), blockRef.BlockHash())))
//}
//
//func (km *mockKeyManager) SignViewChange(vcHeader lh.ViewChangeHeader) lh.SenderSignature {
//	return NewMockSenderSignature(km.MyPublicKey(),
//		lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d", vcHeader.MessageType(), PRIVATE_KEY_PREFIX, km.MyPublicKey(), vcHeader.BlockHeight(), vcHeader.View())))
//}
//
//func (km *mockKeyManager) SignNewView(nvHeader lh.NewViewHeader) lh.SenderSignature {
//	return NewMockSenderSignature(km.MyPublicKey(),
//		lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d", nvHeader.MessageType(), PRIVATE_KEY_PREFIX, km.MyPublicKey(), nvHeader.BlockHeight(), nvHeader.View())))
//}
//
//func (km *mockKeyManager) VerifyBlockRef(blockRef lh.BlockRef, sender lh.SenderSignature) bool {
//
//	if myIdRejected(sender.SenderPublicKey(), km.RejectedPublicKeys) {
//		return false
//	}
//
//	signedMessage := lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d|%s", blockRef.MessageType(), PRIVATE_KEY_PREFIX, sender.SenderPublicKey(), blockRef.BlockHeight(), blockRef.View(), blockRef.BlockHash()))
//	return signedMessage.Equals(sender.Signature())
//}
//
//func (km *mockKeyManager) VerifyViewChange(vcHeader lh.ViewChangeHeader, sender lh.SenderSignature) bool {
//	if myIdRejected(sender.SenderPublicKey(), km.RejectedPublicKeys) {
//		return false
//	}
//
//	signedMessage := lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d", vcHeader.MessageType(), PRIVATE_KEY_PREFIX, sender.SenderPublicKey(), vcHeader.BlockHeight(), vcHeader.View()))
//	return signedMessage.Equals(sender.Signature())
//}
//
//func (km *mockKeyManager) VerifyNewView(nvHeader lh.NewViewHeader, sender lh.SenderSignature) bool {
//	if myIdRejected(sender.SenderPublicKey(), km.RejectedPublicKeys) {
//		return false
//	}
//
//	signedMessage := lh.Signature(fmt.Sprintf("%s|%s|%s|%d|%d", nvHeader.MessageType(), PRIVATE_KEY_PREFIX, sender.SenderPublicKey(), nvHeader.BlockHeight(), nvHeader.View()))
//	return signedMessage.Equals(sender.Signature())
//}

func myIdRejected(id lh.PublicKey, rejected []lh.PublicKey) bool {
	for _, rejectedKey := range rejected {
		if rejectedKey.Equals(id) {
			return true
		}
	}
	return false
}
