package keymanagermock

import (
	"fmt"
	"strings"
)

// TODO Keys should not be strings - convert to our primitives

const PRIVATE_KEY_PREFIX = "PRIVATE_KEY"

type KeyManagerMock struct {
	myPublicKey        []byte
	RejectedPublicKeys [][]byte
}

type KeyManager interface {
	Sign(object []byte) string
	Verify(object []byte, signature string, publicKey []byte) bool
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

func (km *KeyManagerMock) Sign(object []byte) string {
	return fmt.Sprintf("%s-%s-%s", PRIVATE_KEY_PREFIX, km.MyPublicKey, string(object))
}

func (km *KeyManagerMock) Verify(object []byte, signature string, publicKey []byte) bool {
	if IndexOf(km.RejectedPublicKeys, publicKey) > -1 {
		return false
	}

	if !strings.Contains(signature, PRIVATE_KEY_PREFIX) {
		return false
	}

	withoutPrefix := signature[len(PRIVATE_KEY_PREFIX)+1:]
	if !strings.Contains(withoutPrefix, string(publicKey)) {
		return false
	}

	withoutPublicKey := withoutPrefix[len(string(publicKey))+1:]

	if string(object) != withoutPublicKey {
		return false
	}

	// TODO How to convert this to GO??
	//if (JSON.stringify(object) !== withoutPublicKey) {
	//	return false;
	//}

	return true
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
