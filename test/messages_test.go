package test

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildAndReadPreprepareMessage(t *testing.T) {

	// Create PPM

	b1 := builders.CreateBlock(builders.GenesisBlock)
	b2 := builders.CreateBlock(b1)
	b3 := builders.CreateBlock(b2)
	b4 := builders.CreateBlock(b3)

	blocks := []lh.Block{b1, b2, b3, b4}

	mockBlockUtils := builders.NewMockBlockUtils(blocks)
	mockKeyManager := builders.NewMockKeyManager(primitives.Ed25519PublicKey("PK"), nil)
	//mockNetComm := builders.NewMockNetworkCommunication()

	mf := &lh.MessageFactoryImpl{
		BlockUtils: mockBlockUtils,
		KeyManager: mockKeyManager,
	}

	ppm := mf.CreatePreprepareMessage(10, 20, b1)
	ppmBytes := ppm.Raw()

	/*
		This is kinda pointless, not sending and receiving anything
		err := mockNetComm.Send([]lh.PublicKey{lh.PublicKey("PK2")}, ppmBytes)
		if err != nil {
			t.Fatal("failed to send message", err)
		}

		receiver := builders.NewMockMessageReceiver()
		receiver.When("OnReceive", mock.Any).Return(ppmBytes)
	*/

	receivedPPMC := lh.PreprepareMessageContentReader(ppmBytes)

	require.Equal(t, receivedPPMC.SignedHeader().MessageType(), lh.LEAN_HELIX_PREPREPARE, "Message type should be LEAN_HELIX_PREPREPARE")
	require.Equal(t, receivedPPMC.SignedHeader().BlockHeight(), 10, "Height = 10")
	require.Equal(t, receivedPPMC.SignedHeader().View(), 2, "View = 20")

}
