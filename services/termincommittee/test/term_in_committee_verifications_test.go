// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPreprepareVerification(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		h.setNode1AsTheLeader(ctx, 1, 1, block1)

		// start with 0 preprepare
		hasPreprepare := h.hasPreprepare(2, 1, block2)
		require.False(t, hasPreprepare, "No preprepare should exist in the storage")

		// sending a preprepare (height 2)
		h.receiveAndHandlePreprepare(ctx, 1, 2, 1, block2)

		// Expect the storage to have it
		hasPreprepare = h.hasPreprepare(2, 1, block2)
		require.True(t, hasPreprepare, "A preprepare should exist in the storage")

		// sending another preprepare (height 3)
		h.failFutureVerifications()
		h.receiveAndHandlePreprepare(ctx, 1, 3, 1, block3)

		// Expect the storage NOT to have it
		hasPreprepare = h.hasPreprepare(3, 1, block3)
		require.False(t, hasPreprepare, "preprepare should NOT exist in the storage")
	})
}

func TestPrepareVerification(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block := mocks.ABlock(interfaces.GenesisBlock)

		// start with 0 prepare
		prepareCount := h.countPrepare(1, 0, block)
		require.Equal(t, 0, prepareCount, "No prepare should exist in the storage")

		// sending a prepare
		h.receiveAndHandlePrepare(ctx, 1, 1, 0, block)

		// Expect the storage to have it
		prepareCount = h.countPrepare(1, 0, block)
		require.Equal(t, 1, prepareCount, "1 prepare should exist in the storage")

		// sending another (Bad) prepare (From a different node)
		h.failFutureVerifications()
		h.receiveAndHandlePrepare(ctx, 2, 1, 0, block)

		// Expect the storage NOT to store it
		prepareCount = h.countPrepare(1, 0, block)
		require.Equal(t, 1, prepareCount, "(Still) 1 prepare should exist in the storage")
	})
}

func TestViewChangeVerification(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block := mocks.ABlock(interfaces.GenesisBlock)

		// start with 0 view-change
		viewChangeCountOnView4 := h.countViewChange(1, 4)
		viewChangeCountOnView8 := h.countViewChange(1, 8)
		require.Equal(t, 0, viewChangeCountOnView4, "No view-change should exist in the storage, on view 4")
		require.Equal(t, 0, viewChangeCountOnView8, "No view-change should exist in the storage, on view 8")

		// sending a view-change
		h.receiveAndHandleViewChange(ctx, 3, 1, 4, block)

		// Expect the storage to have it
		viewChangeCountOnView4 = h.countViewChange(1, 4)
		require.Equal(t, 1, viewChangeCountOnView4, "1 view-change should exist in the storage, on view 4")
		require.Equal(t, 0, viewChangeCountOnView8, "No view-change should exist in the storage, on view 8")

		// sending another (Bad) view-change
		h.failFutureVerifications()
		h.receiveAndHandleViewChange(ctx, 3, 2, 8, block)

		// Expect the storage NOT to store it
		viewChangeCountOnView4 = h.countViewChange(1, 4)
		viewChangeCountOnView8 = h.countViewChange(1, 8)
		require.Equal(t, 1, viewChangeCountOnView4, "1 view-change should exist in the storage, on view 4")
		require.Equal(t, 0, viewChangeCountOnView8, "(Still) No view-change should exist in the storage, on view 8")
	})
}

func TestNewViewVerification(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness(ctx, t)

		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		h.setNode1AsTheLeader(ctx, 1, 1, block1)

		// start with 0 new-view (Counting the preprepare)
		hasPreprepare := h.hasPreprepare(1, 1, block2)
		require.False(t, hasPreprepare, "No preprepare should exist in the storage")

		// sending a new-view
		h.receiveAndHandleNewView(ctx, 0, 1, 4, block2)

		// Expect the storage to have it
		hasPreprepare = h.hasPreprepare(1, 4, block2)
		require.True(t, hasPreprepare, "A preprepare should exist in the storage")

		// sending another (Bad) new-view
		h.failFutureVerifications()
		h.receiveAndHandleNewView(ctx, 0, 1, 8, block3)

		// Expect the storage to have it
		hasPreprepare = h.hasPreprepare(1, 8, block3)
		require.False(t, hasPreprepare, "A preprepare should NOT exist in the storage")
	})
}
