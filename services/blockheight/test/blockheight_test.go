// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommitteeQuorum(t *testing.T) {
	require.Equal(t, 3, quorum.CalcQuorumSize(4))
	require.Equal(t, 4, quorum.CalcQuorumSize(5))
	require.Equal(t, 5, quorum.CalcQuorumSize(6))
	require.Equal(t, 5, quorum.CalcQuorumSize(7))
	require.Equal(t, 6, quorum.CalcQuorumSize(8))
	require.Equal(t, 7, quorum.CalcQuorumSize(9))
	require.Equal(t, 7, quorum.CalcQuorumSize(10))
	require.Equal(t, 8, quorum.CalcQuorumSize(11))
	require.Equal(t, 9, quorum.CalcQuorumSize(12))
	require.Equal(t, 15, quorum.CalcQuorumSize(21))
	require.Equal(t, 15, quorum.CalcQuorumSize(22))
	require.Equal(t, 67, quorum.CalcQuorumSize(100))
}
