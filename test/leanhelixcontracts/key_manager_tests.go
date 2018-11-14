package leanhelixcontracts

import (
	"github.com/orbs-network/lean-helix-go"
	"testing"
)

// See README.md in this folder for how to invoke this from the consumer of this library
func TestSignAndVerify(t *testing.T, mgr leanhelix.KeyManager) {

	input := []byte("Hello world")
	sig := mgr.Sign(input)
	sender := (&leanhelix.SenderSignatureBuilder{
		SenderPublicKey: mgr.MyPublicKey(),
		Signature:       sig,
	}).Build()

	if !mgr.Verify(input, sender) {
		t.Errorf("Verify failed! input=%v sig=%v pubkey=%v", input, sig, mgr.MyPublicKey())
	}
}
