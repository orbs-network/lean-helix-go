package test

import (
	"bytes"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestPreparedMessagesExtractor(t *testing.T) {
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	leaderId := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	leaderKeyManager := builders.NewMockKeyManager(primitives.MemberId(leaderId))
	sender1KeyManager := builders.NewMockKeyManager(primitives.MemberId(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(primitives.MemberId(senderId2))

	t.Run("should return the prepare proof", func(t *testing.T) {
		ppm := builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, view, block)
		pm1 := builders.APrepareMessage(sender1KeyManager, senderId1, blockHeight, view, block)
		pm2 := builders.APrepareMessage(sender2KeyManager, senderId2, blockHeight, view, block)
		storage := leanhelix.NewInMemoryStorage()
		storage.StorePreprepare(ppm)
		storage.StorePrepare(pm1)
		storage.StorePrepare(pm2)

		expectedProof := &leanhelix.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   []*leanhelix.PrepareMessage{pm1, pm2},
		}

		q := 3

		xpp := expectedProof.PreprepareMessage.Raw()
		xp0 := expectedProof.PrepareMessages[0].Raw()
		xp1 := expectedProof.PrepareMessages[1].Raw()

		actualProof := leanhelix.ExtractPreparedMessages(blockHeight, storage, q)
		app := actualProof.PreprepareMessage.Raw()
		ap0 := actualProof.PrepareMessages[0].Raw()
		ap1 := actualProof.PrepareMessages[1].Raw()

		require.True(t, bytes.Compare(app, xpp) == 0)
		require.True(t, bytes.Compare(ap0, xp0) == 0 || bytes.Compare(ap0, xp1) == 0)
		require.True(t, bytes.Compare(ap1, xp0) == 0 || bytes.Compare(ap1, xp1) == 0)
	})

	t.Run("should return the latest (highest view) Prepare Proof", func(t *testing.T) {
		storage := leanhelix.NewInMemoryStorage()
		ppm10 := builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, 10, block)
		pm10a := builders.APrepareMessage(sender1KeyManager, senderId1, blockHeight, 10, block)
		pm10b := builders.APrepareMessage(sender2KeyManager, senderId2, blockHeight, 10, block)

		ppm20 := builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, 20, block)
		pm20a := builders.APrepareMessage(sender1KeyManager, senderId1, blockHeight, 20, block)
		pm20b := builders.APrepareMessage(sender2KeyManager, senderId2, blockHeight, 20, block)

		ppm30 := builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, 30, block)
		pm30a := builders.APrepareMessage(sender1KeyManager, senderId1, blockHeight, 30, block)
		pm30b := builders.APrepareMessage(sender2KeyManager, senderId2, blockHeight, 30, block)

		storage.StorePreprepare(ppm10)
		storage.StorePrepare(pm10a)
		storage.StorePrepare(pm10b)

		storage.StorePreprepare(ppm20)
		storage.StorePrepare(pm20a)
		storage.StorePrepare(pm20b)

		storage.StorePreprepare(ppm30)
		storage.StorePrepare(pm30a)
		storage.StorePrepare(pm30b)

		expectedProof := &leanhelix.PreparedMessages{
			PreprepareMessage: ppm30,
			PrepareMessages:   []*leanhelix.PrepareMessage{pm30a, pm30b},
		}
		q := 3

		xpp := expectedProof.PreprepareMessage.Raw()
		xp0 := expectedProof.PrepareMessages[0].Raw()
		xp1 := expectedProof.PrepareMessages[1].Raw()

		actualProof := leanhelix.ExtractPreparedMessages(blockHeight, storage, q)
		app := actualProof.PreprepareMessage.Raw()
		ap0 := actualProof.PrepareMessages[0].Raw()
		ap1 := actualProof.PrepareMessages[1].Raw()

		require.True(t, bytes.Compare(app, xpp) == 0)
		require.True(t, bytes.Compare(ap0, xp0) == 0 || bytes.Compare(ap0, xp1) == 0)
		require.True(t, bytes.Compare(ap1, xp0) == 0 || bytes.Compare(ap1, xp1) == 0)
	})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		pm1 := builders.APrepareMessage(sender1KeyManager, senderId1, blockHeight, view, block)
		pm2 := builders.APrepareMessage(sender2KeyManager, senderId2, blockHeight, view, block)
		storage := leanhelix.NewInMemoryStorage()
		storage.StorePrepare(pm1)
		storage.StorePrepare(pm2)
		q := 3
		actualPreparedMessages := leanhelix.ExtractPreparedMessages(blockHeight, storage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		ppm := builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, view, block)
		storage := leanhelix.NewInMemoryStorage()
		storage.StorePreprepare(ppm)
		q := 3
		actualPreparedMessages := leanhelix.ExtractPreparedMessages(blockHeight, storage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfNotEnoughPrepares", func(t *testing.T) {
		ppm := builders.APreprepareMessage(leaderKeyManager, leaderId, blockHeight, view, block)
		pm1 := builders.APrepareMessage(sender1KeyManager, senderId1, blockHeight, view, block)
		storage := leanhelix.NewInMemoryStorage()
		storage.StorePreprepare(ppm)
		storage.StorePrepare(pm1)
		q := 3
		actualPreparedMessages := leanhelix.ExtractPreparedMessages(blockHeight, storage, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if not enough Prepares in storage (# Prepares < 2*f)")
	})
}
