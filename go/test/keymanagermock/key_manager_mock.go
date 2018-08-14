package keymanagermock

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/go/networkcommunication"
)

// TODO Keys should not be strings - convert to our primitives

const PRIVATE_KEY_PREFIX = "PRIVATE_KEY"

type KeyManagerMock struct {
	myPublicKey        []byte
	RejectedPublicKeys [][]byte
}

type KeyManager interface {
	Sign(ppd *networkcommunication.PrepreparePayloadData) string
	Verify(ppd *networkcommunication.PrepreparePayloadData, signature string, publicKey []byte) bool
	MyPublicKey() []byte
}

func NewKeyManagerMock(publicKey []byte, rejectedPublicKeys [][]byte) *KeyManagerMock {
	return &KeyManagerMock{
		myPublicKey:        publicKey,
		RejectedPublicKeys: rejectedPublicKeys,
	}
}

func (km *KeyManagerMock) MyPublicKey() []byte {
	return km.myPublicKey
}

func (km *KeyManagerMock) Sign(ppd *networkcommunication.PrepreparePayloadData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", PRIVATE_KEY_PREFIX, km.MyPublicKey(), string(ppd.Term), string(ppd.View), string(ppd.BlockHash))
}

func (km *KeyManagerMock) Verify(ppd *networkcommunication.PrepreparePayloadData, signature string, publicKey []byte) bool {
	if IndexOf(km.RejectedPublicKeys, publicKey) > -1 {
		return false
	}

	expectedSignature := fmt.Sprintf("%s|%s|%s|%s|%s", PRIVATE_KEY_PREFIX, publicKey, string(ppd.Term), string(ppd.View), string(ppd.BlockHash))

	return expectedSignature == signature
}

//TODO Find a Go way to compare []byte's
func IndexOf(publicKeys [][]byte, searchTerm []byte) int {
	if publicKeys == nil {
		return -1
	}
	for i := 0; i < len(publicKeys); i++ {
		if string(searchTerm) == string(publicKeys[i]) {
			return i
		}
	}
	return -1
}
