package test

// TODO incomplete

import (
	"bytes"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestPreparedMessagesExtractor(t *testing.T) {

	//const logger: Logger = new SilentLogger();
	height := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	leaderId := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	leaderKeyManager := builders.NewMockKeyManager(Ed25519PublicKey(leaderId))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))
	leaderMsgFactory := lh.NewMessageFactory(leaderKeyManager)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)

	//f := 1

	t.Run("should return the prepare proof", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		ppm := leaderMsgFactory.CreatePreprepareMessage(height, view, block)
		pm1 := sender1MsgFactory.CreatePrepareMessage(height, view, blockHash)
		pm2 := sender2MsgFactory.CreatePrepareMessage(height, view, blockHash)
		myStorage.StorePreprepare(ppm)
		myStorage.StorePrepare(pm1)
		myStorage.StorePrepare(pm2)

		expectedProof := &lh.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   []*lh.PrepareMessage{pm1, pm2},
		}

		q := 3

		actualProof := lh.ExtractPreparedMessages(height, myStorage, q)
		require.True(t, bytes.Compare(expectedProof.PreprepareMessage.Raw(), actualProof.PreprepareMessage.Raw()) == 0)
		require.True(t, bytes.Compare(expectedProof.PrepareMessages[0].Raw(), actualProof.PrepareMessages[0].Raw()) == 0)
		require.True(t, bytes.Compare(expectedProof.PrepareMessages[1].Raw(), actualProof.PrepareMessages[1].Raw()) == 0)

	})

	t.Run("should return the latest (highest view) Preprepare message", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		ppm10 := leaderMsgFactory.CreatePreprepareMessage(height, 10, block)
		ppm20 := leaderMsgFactory.CreatePreprepareMessage(height, 20, block)
		ppm30 := leaderMsgFactory.CreatePreprepareMessage(height, 30, block)
		myStorage.StorePreprepare(ppm10)
		myStorage.StorePreprepare(ppm20)
		myStorage.StorePreprepare(ppm30)

		actualPPM, _ := myStorage.GetLatestPreprepare(height)
		require.Equal(t, actualPPM.View(), View(30), "View of Preprepare message should be 30 (highest for this block height)")
	})

	t.Run("should return the latest (highest view) Prepare Proof", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		ppm10 := leaderMsgFactory.CreatePreprepareMessage(height, 10, block)
		pm10a := sender1MsgFactory.CreatePrepareMessage(height, 10, blockHash)
		pm10b := sender2MsgFactory.CreatePrepareMessage(height, 10, blockHash)

		ppm20 := leaderMsgFactory.CreatePreprepareMessage(height, 20, block)
		pm20a := sender1MsgFactory.CreatePrepareMessage(height, 20, blockHash)
		pm20b := sender2MsgFactory.CreatePrepareMessage(height, 20, blockHash)

		ppm30 := leaderMsgFactory.CreatePreprepareMessage(height, 30, block)
		pm30a := sender1MsgFactory.CreatePrepareMessage(height, 30, blockHash)
		pm30b := sender2MsgFactory.CreatePrepareMessage(height, 30, blockHash)

		myStorage.StorePreprepare(ppm10)
		myStorage.StorePrepare(pm10a)
		myStorage.StorePrepare(pm10b)

		myStorage.StorePreprepare(ppm20)
		myStorage.StorePrepare(pm20a)
		myStorage.StorePrepare(pm20b)

		myStorage.StorePreprepare(ppm30)
		myStorage.StorePrepare(pm30a)
		myStorage.StorePrepare(pm30b)

		expectedPreparedMessages := &lh.PreparedMessages{
			PreprepareMessage: ppm30,
			PrepareMessages:   []*lh.PrepareMessage{pm30a, pm30b},
		}
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(height, myStorage, q)
		require.True(t, bytes.Compare(expectedPreparedMessages.PreprepareMessage.Raw(), actualPreparedMessages.PreprepareMessage.Raw()) == 0)
		require.True(t, bytes.Compare(expectedPreparedMessages.PrepareMessages[0].Raw(), actualPreparedMessages.PrepareMessages[0].Raw()) == 0)
		require.True(t, bytes.Compare(expectedPreparedMessages.PrepareMessages[1].Raw(), actualPreparedMessages.PrepareMessages[1].Raw()) == 0)

	})

	// TODO This "TestStoreAndGetPrepareProof" test will always PASS if "TestReturnPreparedProofWithHighestView" below passes, consider deleting
	//t.Run("TestStoreAndGetPrepareProof", func(t *testing.T) {
	//	myStorage := lh.NewInMemoryStorage()
	//	myStorage.StorePreprepare(ppm)
	//	myStorage.StorePrepare(pm2)
	//	myStorage.StorePrepare(pm1)
	//	expectedProof := lh.CreatePreparedProof(ppm, []lh.PrepareMessage{pm1, pm2})
	//
	//	actualProof, _ := myStorage.GetLatestPrepared(height, f)
	//	compPrepareProof(t, actualProof, expectedProof, "return a prepared proof generated by the PPM and PMs in storage")
	//})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		pm1 := sender1MsgFactory.CreatePrepareMessage(height, view, blockHash)
		pm2 := sender2MsgFactory.CreatePrepareMessage(height, view, blockHash)
		myStorage := lh.NewInMemoryStorage()
		myStorage.StorePrepare(pm1)
		myStorage.StorePrepare(pm2)
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(height, myStorage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		ppm := leaderMsgFactory.CreatePreprepareMessage(height, view, block)
		myStorage := lh.NewInMemoryStorage()
		myStorage.StorePreprepare(ppm)
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(height, myStorage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfNotEnoughPrepares", func(t *testing.T) {
		ppm := leaderMsgFactory.CreatePreprepareMessage(height, view, block)
		pm1 := sender1MsgFactory.CreatePrepareMessage(height, view, blockHash)
		myStorage := lh.NewInMemoryStorage()
		myStorage.StorePreprepare(ppm)
		myStorage.StorePrepare(pm1)
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(height, myStorage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if not enough Prepares in storage (# Prepares < 2*f)")
	})
}

// TODO GetLatestPrepared() should initially be here as in TS code but later moved out, because it contains algo logic (it checks something with 2*f))
