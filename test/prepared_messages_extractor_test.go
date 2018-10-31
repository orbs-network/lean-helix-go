package test

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
	blockHeight := BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := View(math.Floor(rand.Float64() * 1000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	leaderId := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	leaderKeyManager := builders.NewMockKeyManager(Ed25519PublicKey(leaderId))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))

	t.Run("should return the prepare proof", func(t *testing.T) {
		ppm := builders.APreprepareMessage(leaderKeyManager, blockHeight, view, block)
		pm1 := builders.APrepareMessage(sender1KeyManager, blockHeight, view, block)
		pm2 := builders.APrepareMessage(sender2KeyManager, blockHeight, view, block)
		storage := lh.NewInMemoryStorage()
		storage.StorePreprepare(ppm)
		storage.StorePrepare(pm1)
		storage.StorePrepare(pm2)

		expectedProof := &lh.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   []*lh.PrepareMessage{pm1, pm2},
		}

		q := 3

		actualProof := lh.ExtractPreparedMessages(blockHeight, storage, q)
		require.True(t, bytes.Compare(expectedProof.PreprepareMessage.Raw(), actualProof.PreprepareMessage.Raw()) == 0)
		require.True(t, bytes.Compare(expectedProof.PrepareMessages[0].Raw(), actualProof.PrepareMessages[0].Raw()) == 0)
		require.True(t, bytes.Compare(expectedProof.PrepareMessages[1].Raw(), actualProof.PrepareMessages[1].Raw()) == 0)

	})

	t.Run("should return the latest (highest view) Prepare Proof", func(t *testing.T) {
		storage := lh.NewInMemoryStorage()
		ppm10 := builders.APreprepareMessage(leaderKeyManager, blockHeight, 10, block)
		pm10a := builders.APrepareMessage(sender1KeyManager, blockHeight, 10, block)
		pm10b := builders.APrepareMessage(sender2KeyManager, blockHeight, 10, block)

		ppm20 := builders.APreprepareMessage(leaderKeyManager, blockHeight, 20, block)
		pm20a := builders.APrepareMessage(sender1KeyManager, blockHeight, 20, block)
		pm20b := builders.APrepareMessage(sender2KeyManager, blockHeight, 20, block)

		ppm30 := builders.APreprepareMessage(leaderKeyManager, blockHeight, 30, block)
		pm30a := builders.APrepareMessage(sender1KeyManager, blockHeight, 30, block)
		pm30b := builders.APrepareMessage(sender2KeyManager, blockHeight, 30, block)

		storage.StorePreprepare(ppm10)
		storage.StorePrepare(pm10a)
		storage.StorePrepare(pm10b)

		storage.StorePreprepare(ppm20)
		storage.StorePrepare(pm20a)
		storage.StorePrepare(pm20b)

		storage.StorePreprepare(ppm30)
		storage.StorePrepare(pm30a)
		storage.StorePrepare(pm30b)

		expectedPreparedMessages := &lh.PreparedMessages{
			PreprepareMessage: ppm30,
			PrepareMessages:   []*lh.PrepareMessage{pm30a, pm30b},
		}
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(blockHeight, storage, q)
		require.True(t, bytes.Compare(expectedPreparedMessages.PreprepareMessage.Raw(), actualPreparedMessages.PreprepareMessage.Raw()) == 0)
		require.True(t, bytes.Compare(expectedPreparedMessages.PrepareMessages[0].Raw(), actualPreparedMessages.PrepareMessages[0].Raw()) == 0)
		require.True(t, bytes.Compare(expectedPreparedMessages.PrepareMessages[1].Raw(), actualPreparedMessages.PrepareMessages[1].Raw()) == 0)

	})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		pm1 := builders.APrepareMessage(sender1KeyManager, blockHeight, view, block)
		pm2 := builders.APrepareMessage(sender2KeyManager, blockHeight, view, block)
		storage := lh.NewInMemoryStorage()
		storage.StorePrepare(pm1)
		storage.StorePrepare(pm2)
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(blockHeight, storage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		ppm := builders.APreprepareMessage(leaderKeyManager, blockHeight, view, block)
		storage := lh.NewInMemoryStorage()
		storage.StorePreprepare(ppm)
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(blockHeight, storage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfNotEnoughPrepares", func(t *testing.T) {
		ppm := builders.APreprepareMessage(leaderKeyManager, blockHeight, view, block)
		pm1 := builders.APrepareMessage(sender1KeyManager, blockHeight, view, block)
		storage := lh.NewInMemoryStorage()
		storage.StorePreprepare(ppm)
		storage.StorePrepare(pm1)
		q := 3
		actualPreparedMessages := lh.ExtractPreparedMessages(blockHeight, storage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if not enough Prepares in storage (# Prepares < 2*f)")
	})
}
