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

//func TestBuildAndReadPrepareMessage(t *testing.T) {
//	b1 := builders.CreateBlock(builders.GenesisBlock)
//	mockKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK"), nil)
//	mf := &lh.MessageFactoryImpl{
//		KeyManager: mockKeyManager,
//	}
//	pm := mf.CreatePrepareMessage(11, 21, b1.BlockHash())
//	pmBytes := pm.Raw()
//	receivedPMC := lh.PrepareMessageContentReader(pmBytes)
//	require.Equal(t, receivedPMC.SignedHeader().MessageType(), lh.LEAN_HELIX_PREPARE, "Message type should be LEAN_HELIX_PREPARE")
//	require.True(t, receivedPMC.SignedHeader().BlockHeight().Equal(11), "Height = 11")
//	require.True(t, receivedPMC.SignedHeader().View().Equal(21), "View = 21")
//}
//
//func TestBuildAndReadCommitMessage(t *testing.T) {
//	b1 := builders.CreateBlock(builders.GenesisBlock)
//	mockKeyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK"), nil)
//	mf := &lh.MessageFactoryImpl{
//		KeyManager: mockKeyManager,
//	}
//	cm := mf.CreateCommitMessage(12, 22, b1.BlockHash())
//	cmBytes := cm.Raw()
//	receivedCMC := lh.CommitMessageContentReader(cmBytes)
//	require.Equal(t, receivedCMC.SignedHeader().MessageType(), lh.LEAN_HELIX_COMMIT, "Message type should be LEAN_HELIX_COMMIT")
//	require.True(t, receivedCMC.SignedHeader().BlockHeight().Equal(12), "Height = 12")
//	require.True(t, receivedCMC.SignedHeader().View().Equal(22), "View = 22")
//}
