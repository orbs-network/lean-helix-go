// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"bytes"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/testhelpers"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestPreparedMessagesExtractor(t *testing.T) {
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Intn(10))
	block := mocks.ABlock(interfaces.GenesisBlock)
	leaderIdW1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderIdW2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderIdW3 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderIdW4 := primitives.MemberId(strconv.Itoa(rand.Int()))

	ids := []primitives.MemberId{leaderIdW1, senderIdW2, senderIdW3, senderIdW4}
	weights := []uint64{1, 2, 3, 4}
	committeeMembers := testhelpers.GenMembersWithWeights(ids, weights)

	require.Equal(t, uint(7), quorum.CalcQuorumWeight(quorum.GetWeights(committeeMembers)))

	leaderW1KeyManager := mocks.NewMockKeyManager(primitives.MemberId(leaderIdW1))
	senderW2KeyManager := mocks.NewMockKeyManager(primitives.MemberId(senderIdW2))
	senderW3KeyManager := mocks.NewMockKeyManager(primitives.MemberId(senderIdW3))
	senderW4KeyManager := mocks.NewMockKeyManager(primitives.MemberId(senderIdW4))

	t.Run("should return the prepare proof", func(t *testing.T) {
		ppm := builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, view, block)
		pm1 := builders.APrepareMessage(instanceId, senderW2KeyManager, senderIdW2, blockHeight, view, block)
		pm2 := builders.APrepareMessage(instanceId, senderW4KeyManager, senderIdW4, blockHeight, view, block)
		s := storage.NewInMemoryStorage()
		s.StorePreprepare(ppm)
		s.StorePrepare(pm1)
		s.StorePrepare(pm2)

		expectedProof := &preparedmessages.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   []*interfaces.PrepareMessage{pm1, pm2},
		}

		xpp := expectedProof.PreprepareMessage.Raw()
		xp0 := expectedProof.PrepareMessages[0].Raw()
		xp1 := expectedProof.PrepareMessages[1].Raw()

		actualProof := preparedmessages.ExtractPreparedMessages(blockHeight, view, s, committeeMembers)
		app := actualProof.PreprepareMessage.Raw()
		ap0 := actualProof.PrepareMessages[0].Raw()
		ap1 := actualProof.PrepareMessages[1].Raw()

		require.True(t, bytes.Compare(app, xpp) == 0)
		require.True(t, bytes.Compare(ap0, xp0) == 0 || bytes.Compare(ap0, xp1) == 0)
		require.True(t, bytes.Compare(ap1, xp0) == 0 || bytes.Compare(ap1, xp1) == 0)
	})

	t.Run("should return the latest (highest view) Prepare Proof", func(t *testing.T) {
		s := storage.NewInMemoryStorage()
		ppm10 := builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, 10, block)
		pm10a := builders.APrepareMessage(instanceId, senderW2KeyManager, senderIdW2, blockHeight, 10, block)
		pm10b := builders.APrepareMessage(instanceId, senderW4KeyManager, senderIdW4, blockHeight, 10, block)

		ppm20 := builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, 20, block)
		pm20a := builders.APrepareMessage(instanceId, senderW2KeyManager, senderIdW2, blockHeight, 20, block)
		pm20b := builders.APrepareMessage(instanceId, senderW4KeyManager, senderIdW4, blockHeight, 20, block)

		ppm30 := builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, 30, block)
		pm30a := builders.APrepareMessage(instanceId, senderW2KeyManager, senderIdW2, blockHeight, 30, block)
		pm30b := builders.APrepareMessage(instanceId, senderW4KeyManager, senderIdW4, blockHeight, 30, block)

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

		xpp := expectedProof.PreprepareMessage.Raw()
		xp0 := expectedProof.PrepareMessages[0].Raw()
		xp1 := expectedProof.PrepareMessages[1].Raw()

		actualProof := preparedmessages.ExtractPreparedMessages(blockHeight, 30, s, committeeMembers)
		app := actualProof.PreprepareMessage.Raw()
		ap0 := actualProof.PrepareMessages[0].Raw()
		ap1 := actualProof.PrepareMessages[1].Raw()

		require.True(t, bytes.Compare(app, xpp) == 0)
		require.True(t, bytes.Compare(ap0, xp0) == 0 || bytes.Compare(ap0, xp1) == 0)
		require.True(t, bytes.Compare(ap1, xp0) == 0 || bytes.Compare(ap1, xp1) == 0)
	})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		pm1 := builders.APrepareMessage(instanceId, senderW2KeyManager, senderIdW3, blockHeight, view, block)
		pm2 := builders.APrepareMessage(instanceId, senderW4KeyManager, senderIdW4, blockHeight, view, block)

		// Quorum is reached, but no preprepare msg present

		s := storage.NewInMemoryStorage()
		s.StorePrepare(pm1)
		s.StorePrepare(pm2)
		actualPreparedMessages := preparedmessages.ExtractPreparedMessages(blockHeight, view, s, committeeMembers)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		ppm := builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, view, block)
		s := storage.NewInMemoryStorage()
		s.StorePreprepare(ppm)
		actualPreparedMessages := preparedmessages.ExtractPreparedMessages(blockHeight, view, s, committeeMembers)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfPreparesHaventReachedQuorum", func(t *testing.T) {
		ppm := builders.APreprepareMessage(instanceId, leaderW1KeyManager, leaderIdW1, blockHeight, view, block)
		pm1 := builders.APrepareMessage(instanceId, senderW2KeyManager, senderIdW2, blockHeight, view, block)
		pm2 := builders.APrepareMessage(instanceId, senderW3KeyManager, senderIdW3, blockHeight, view, block)

		// Not reaching the quorum weight - 7

		s := storage.NewInMemoryStorage()
		s.StorePreprepare(ppm)
		s.StorePrepare(pm1)
		s.StorePrepare(pm2)
		actualPreparedMessages := preparedmessages.ExtractPreparedMessages(blockHeight, view, s, committeeMembers)
		require.Nil(t, actualPreparedMessages, "Don't return PreparedMessages from latest view if not enough Prepares in storage (# Prepares < 2*f)")
	})
}
