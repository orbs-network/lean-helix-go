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

func TestCommitteeQuorum(t *testing.T) { // todo more comprehensive
	require.Equal(t, uint(3), quorum.CalcQuorumWeight([]uint{4}))
	require.Equal(t, uint(4), quorum.CalcQuorumWeight([]uint{5}))
	require.Equal(t, uint(5), quorum.CalcQuorumWeight([]uint{6}))
	require.Equal(t, uint(5), quorum.CalcQuorumWeight([]uint{7}))
	require.Equal(t, uint(6), quorum.CalcQuorumWeight([]uint{8}))
	require.Equal(t, uint(7), quorum.CalcQuorumWeight([]uint{9}))
	require.Equal(t, uint(7), quorum.CalcQuorumWeight([]uint{10}))
	require.Equal(t, uint(8), quorum.CalcQuorumWeight([]uint{11}))
	require.Equal(t, uint(9), quorum.CalcQuorumWeight([]uint{12}))
	require.Equal(t, uint(15), quorum.CalcQuorumWeight([]uint{21}))
	require.Equal(t, uint(15), quorum.CalcQuorumWeight([]uint{22}))
	require.Equal(t, uint(67), quorum.CalcQuorumWeight([]uint{100}))
}
