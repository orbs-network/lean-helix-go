package test

import (
	"bytes"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestPreparedMessagesExtractor(t *testing.T) {
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	block := mocks.ABlock(interfaces.GenesisBlock)
	leaderId := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	leaderKeyManager := mocks.NewMockKeyManager(primitives.MemberId(leaderId))
	sender1KeyManager := mocks.NewMockKeyManager(primitives.MemberId(senderId1))
	sender2KeyManager := mocks.NewMockKeyManager(primitives.MemberId(senderId2))

	t.Run("should return the prepare proof", func(t *testing.T) {
		ppm := builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, view, block)
		pm1 := builders.APrepareMessage(instanceId, sender1KeyManager, senderId1, blockHeight, view, block)
		pm2 := builders.APrepareMessage(instanceId, sender2KeyManager, senderId2, blockHeight, view, block)
		s := storage.NewInMemoryStorage()
		s.StorePreprepare(ppm)
		s.StorePrepare(pm1)
		s.StorePrepare(pm2)

		expectedProof := &preparedmessages.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   []*interfaces.PrepareMessage{pm1, pm2},
		}

		q := 3

		xpp := expectedProof.PreprepareMessage.Raw()
		xp0 := expectedProof.PrepareMessages[0].Raw()
		xp1 := expectedProof.PrepareMessages[1].Raw()

		actualProof := preparedmessages.ExtractPreparedMessages(blockHeight, s, q)
		app := actualProof.PreprepareMessage.Raw()
		ap0 := actualProof.PrepareMessages[0].Raw()
		ap1 := actualProof.PrepareMessages[1].Raw()

		require.True(t, bytes.Compare(app, xpp) == 0)
		require.True(t, bytes.Compare(ap0, xp0) == 0 || bytes.Compare(ap0, xp1) == 0)
		require.True(t, bytes.Compare(ap1, xp0) == 0 || bytes.Compare(ap1, xp1) == 0)
	})

	t.Run("should return the latest (highest view) Prepare Proof", func(t *testing.T) {
		s := storage.NewInMemoryStorage()
		ppm10 := builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, 10, block)
		pm10a := builders.APrepareMessage(instanceId, sender1KeyManager, senderId1, blockHeight, 10, block)
		pm10b := builders.APrepareMessage(instanceId, sender2KeyManager, senderId2, blockHeight, 10, block)

		ppm20 := builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, 20, block)
		pm20a := builders.APrepareMessage(instanceId, sender1KeyManager, senderId1, blockHeight, 20, block)
		pm20b := builders.APrepareMessage(instanceId, sender2KeyManager, senderId2, blockHeight, 20, block)

		ppm30 := builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, 30, block)
		pm30a := builders.APrepareMessage(instanceId, sender1KeyManager, senderId1, blockHeight, 30, block)
		pm30b := builders.APrepareMessage(instanceId, sender2KeyManager, senderId2, blockHeight, 30, block)

		s.StorePreprepare(ppm10)
		s.StorePrepare(pm10a)
		s.StorePrepare(pm10b)

		s.StorePreprepare(ppm20)
		s.StorePrepare(pm20a)
		s.StorePrepare(pm20b)

		s.StorePreprepare(ppm30)
		s.StorePrepare(pm30a)
		s.StorePrepare(pm30b)

		expectedProof := &preparedmessages.PreparedMessages{
			PreprepareMessage: ppm30,
			PrepareMessages:   []*interfaces.PrepareMessage{pm30a, pm30b},
		}
		q := 3

		xpp := expectedProof.PreprepareMessage.Raw()
		xp0 := expectedProof.PrepareMessages[0].Raw()
		xp1 := expectedProof.PrepareMessages[1].Raw()

		actualProof := preparedmessages.ExtractPreparedMessages(blockHeight, s, q)
		app := actualProof.PreprepareMessage.Raw()
		ap0 := actualProof.PrepareMessages[0].Raw()
		ap1 := actualProof.PrepareMessages[1].Raw()

		require.True(t, bytes.Compare(app, xpp) == 0)
		require.True(t, bytes.Compare(ap0, xp0) == 0 || bytes.Compare(ap0, xp1) == 0)
		require.True(t, bytes.Compare(ap1, xp0) == 0 || bytes.Compare(ap1, xp1) == 0)
	})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		pm1 := builders.APrepareMessage(instanceId, sender1KeyManager, senderId1, blockHeight, view, block)
		pm2 := builders.APrepareMessage(instanceId, sender2KeyManager, senderId2, blockHeight, view, block)
		s := storage.NewInMemoryStorage()
		s.StorePrepare(pm1)
		s.StorePrepare(pm2)
		q := 3
		actualPreparedMessages := preparedmessages.ExtractPreparedMessages(blockHeight, s, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		ppm := builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, view, block)
		s := storage.NewInMemoryStorage()
		s.StorePreprepare(ppm)
		q := 3
		actualPreparedMessages := preparedmessages.ExtractPreparedMessages(blockHeight, s, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfNotEnoughPrepares", func(t *testing.T) {
		ppm := builders.APreprepareMessage(instanceId, leaderKeyManager, leaderId, blockHeight, view, block)
		pm1 := builders.APrepareMessage(instanceId, sender1KeyManager, senderId1, blockHeight, view, block)
		s := storage.NewInMemoryStorage()
		s.StorePreprepare(ppm)
		s.StorePrepare(pm1)
		q := 3
		actualPreparedMessages := preparedmessages.ExtractPreparedMessages(blockHeight, s, q)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if not enough Prepares in storage (# Prepares < 2*f)")
	})
}
