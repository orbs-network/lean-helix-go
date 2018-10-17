package test

import (
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildAndReadPreprepareMessage(t *testing.T) {
	b1 := builders.CreateBlock(builders.GenesisBlock)
	mockKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK"), nil)
	mf := &lh.MessageFactoryImpl{
		KeyManager: mockKeyManager,
	}
	ppm := mf.CreatePreprepareMessage(10, 20, b1)
	ppmBytes := ppm.Raw()
	receivedPPMC := lh.PreprepareMessageContentReader(ppmBytes)
	require.Equal(t, receivedPPMC.SignedHeader().MessageType(), lh.LEAN_HELIX_PREPREPARE, "Message type should be LEAN_HELIX_PREPREPARE")
	require.True(t, receivedPPMC.SignedHeader().BlockHeight().Equal(10), "Height = 10")
	require.True(t, receivedPPMC.SignedHeader().View().Equal(20), "View = 20")
}
