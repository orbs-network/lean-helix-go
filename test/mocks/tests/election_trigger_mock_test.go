// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package tests

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestElectionTriggerMockInitialization(t *testing.T) {
	actual := mocks.NewMockElectionTrigger()
	require.NotNil(t, actual)
}

func TestCallingCallback(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := mocks.NewMockElectionTrigger()
		var actualView primitives.View = 666
		var actualHeight primitives.BlockHeight = 666
		var expectedView primitives.View = 10
		var expectedHeight primitives.BlockHeight = 20
		cb := func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
			actualHeight = blockHeight
			actualView = view
		}
		et.RegisterOnElection(ctx, expectedHeight, expectedView, cb)

		go et.ManualTrigger(ctx)
		trigger := <-et.ElectionChannel()
		trigger.MoveToNextLeader(ctx)

		require.Equal(t, expectedView, actualView)
		require.Equal(t, expectedHeight, actualHeight)
	})
}

func TestIgnoreEmptyCallback(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := mocks.NewMockElectionTrigger()

		go et.ManualTrigger(ctx)
		trigger := <-et.ElectionChannel()
		trigger.MoveToNextLeader(ctx)
	})
}
