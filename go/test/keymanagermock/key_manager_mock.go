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
	myPublicKey        []byte
	RejectedPublicKeys [][]byte
}

func NewMockKeyManager(publicKey []byte, rejectedPublicKeys [][]byte) *mockKeyManager {
	return &mockKeyManager{
		myPublicKey:        publicKey,
		RejectedPublicKeys: rejectedPublicKeys,
	}
}

func (km *mockKeyManager) MyPublicKey() []byte {
	return km.myPublicKey
}

func (km *mockKeyManager) SignPrepreparePayloadData(ppd *leanhelix.PrepreparePayloadData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", PRIVATE_KEY_PREFIX, km.MyPublicKey(), string(ppd.Term), string(ppd.View), string(ppd.BlockHash))
}
func (km *mockKeyManager) SignPreparePayloadData(pd *leanhelix.PreparePayloadData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", PRIVATE_KEY_PREFIX, km.MyPublicKey(), string(pd.Term), string(pd.View), string(pd.BlockHash))
}
func (km *mockKeyManager) SignCommitPayloadData(cd *leanhelix.CommitPayloadData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", PRIVATE_KEY_PREFIX, km.MyPublicKey(), string(cd.Term), string(cd.View), string(cd.BlockHash))
}

func (km *mockKeyManager) Verify(ppd *leanhelix.PrepreparePayloadData, signature string, publicKey []byte) bool {
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
