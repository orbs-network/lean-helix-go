// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommitteeQuorum(t *testing.T) {
	require.Equal(t, uint(3), quorum.CalcQuorumWeight([]uint64{4}))
	require.Equal(t, uint(4), quorum.CalcQuorumWeight([]uint64{5}))
	require.Equal(t, uint(5), quorum.CalcQuorumWeight([]uint64{6}))
	require.Equal(t, uint(5), quorum.CalcQuorumWeight([]uint64{7}))
	require.Equal(t, uint(6), quorum.CalcQuorumWeight([]uint64{8}))
	require.Equal(t, uint(7), quorum.CalcQuorumWeight([]uint64{9}))
	require.Equal(t, uint(7), quorum.CalcQuorumWeight([]uint64{10}))
	require.Equal(t, uint(8), quorum.CalcQuorumWeight([]uint64{11}))
	require.Equal(t, uint(9), quorum.CalcQuorumWeight([]uint64{12}))
	require.Equal(t, uint(15), quorum.CalcQuorumWeight([]uint64{21}))
	require.Equal(t, uint(15), quorum.CalcQuorumWeight([]uint64{22}))
	require.Equal(t, uint(67), quorum.CalcQuorumWeight([]uint64{100}))

	require.Equal(t, uint(67), quorum.CalcQuorumWeight([]uint64{10, 90}))
	require.Equal(t, uint(67), quorum.CalcQuorumWeight([]uint64{10, 20, 70}))
}

func genCommittee(weights []uint64) []interfaces.CommitteeMember {
	committee := make([]interfaces.CommitteeMember, len(weights))
	for i := 0; i < len(weights); i++ {
		committee[i] = interfaces.CommitteeMember{
			Id:     []byte{byte(i)},
			Weight: weights[i],
		}
	}
	return committee
}

func TestIsQuorum(t *testing.T) {
	committee := genCommittee([]uint64{1, 6, 10, 23, 60})

	ids := func(inds []int) []primitives.MemberId {
		_ids := make([]primitives.MemberId, len(inds))
		for i := 0; i < len(inds); i++ {
			_ids[i] = committee[inds[i]].Id
		}
		return _ids
	}

	isQuorum, totalWeights, q := quorum.IsQuorum(ids([]int{0, 1, 2}), committee)
	require.Equal(t, false, isQuorum)
	require.Equal(t, uint(67), q)
	require.Equal(t, uint(17), totalWeights)

	isQuorum, totalWeights, q = quorum.IsQuorum(ids([]int{2, 3, 4}), committee)
	require.Equal(t, true, isQuorum)
	require.Equal(t, uint(67), q)
	require.Equal(t, uint(93), totalWeights)

	isQuorum, totalWeights, q = quorum.IsQuorum(ids([]int{0, 1, 4}), committee)
	require.Equal(t, true, isQuorum)
	require.Equal(t, uint(67), q)
	require.Equal(t, uint(67), totalWeights)

	isQuorum, totalWeights, q = quorum.IsQuorum(ids([]int{1, 4}), committee)
	require.Equal(t, false, isQuorum)
	require.Equal(t, uint(67), q)
	require.Equal(t, uint(66), totalWeights)

	isQuorum, totalWeights, q = quorum.IsQuorum(ids([]int{4, 4}), committee)
	require.Equal(t, false, isQuorum)
	require.Equal(t, uint(67), q)
	require.Equal(t, uint(60), totalWeights)
}
