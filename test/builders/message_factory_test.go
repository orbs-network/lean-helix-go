package builders

import (
	"bytes"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestBuildAndRead(t *testing.T) {
	keyManager := NewMockKeyManager(Ed25519PublicKey("My PK"))
	height := BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := View(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(GenesisBlock)
	fac := lh.NewMessageFactory(keyManager)

	actualPPM := fac.CreatePreprepareMessage(height, view, block)

	bytes1 := actualPPM.Raw()
	newPPMC := lh.PreprepareMessageContentReader(bytes1)
	bytes2 := newPPMC.Raw()

	require.True(t, bytes.Compare(bytes1, bytes2) == 0)
}

func TestMessageFactory(t *testing.T) {
	keyManager := NewMockKeyManager(Ed25519PublicKey("My PK"))
	height := BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := View(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(GenesisBlock)
	blockHash := block.BlockHash()
	fac := lh.NewMessageFactory(keyManager)

	t.Run("create PreprepareMessage", func(t *testing.T) {
		signedHeader := &lh.BlockRefBuilder{
			MessageType: lh.LEAN_HELIX_PREPREPARE,
			BlockHeight: height,
			View:        view,
			BlockHash:   blockHash,
		}
		ppmcb := &lh.PreprepareMessageContentBuilder{
			SignedHeader: signedHeader,
			Sender: &lh.SenderSignatureBuilder{
				SenderPublicKey: keyManager.MyPublicKey(),
				Signature:       keyManager.Sign(signedHeader.Build().Raw()),
			},
		}

		expectedPPM := &lh.PreprepareMessageImpl{
			Content: ppmcb.Build(),
			MyBlock: block,
		}

		actualPPM := fac.CreatePreprepareMessage(height, view, block)

		require.True(t, bytes.Compare(expectedPPM.Raw(), actualPPM.Raw()) == 0, "compared bytes of PPM")
	})

}
