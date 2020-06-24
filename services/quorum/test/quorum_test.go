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
	require.Equal(t, uint(3), quorum.CalcQuorumWeight([]primitives.MemberWeight{4}))
	require.Equal(t, uint(4), quorum.CalcQuorumWeight([]primitives.MemberWeight{5}))
	require.Equal(t, uint(5), quorum.CalcQuorumWeight([]primitives.MemberWeight{6}))
	require.Equal(t, uint(5), quorum.CalcQuorumWeight([]primitives.MemberWeight{7}))
	require.Equal(t, uint(6), quorum.CalcQuorumWeight([]primitives.MemberWeight{8}))
	require.Equal(t, uint(7), quorum.CalcQuorumWeight([]primitives.MemberWeight{9}))
	require.Equal(t, uint(7), quorum.CalcQuorumWeight([]primitives.MemberWeight{10}))
	require.Equal(t, uint(8), quorum.CalcQuorumWeight([]primitives.MemberWeight{11}))
	require.Equal(t, uint(9), quorum.CalcQuorumWeight([]primitives.MemberWeight{12}))
	require.Equal(t, uint(15), quorum.CalcQuorumWeight([]primitives.MemberWeight{21}))
	require.Equal(t, uint(15), quorum.CalcQuorumWeight([]primitives.MemberWeight{22}))
	require.Equal(t, uint(67), quorum.CalcQuorumWeight([]primitives.MemberWeight{100}))

	require.Equal(t, uint(67), quorum.CalcQuorumWeight([]primitives.MemberWeight{10, 90}))
	require.Equal(t, uint(67), quorum.CalcQuorumWeight([]primitives.MemberWeight{10, 20, 70}))
}

func genCommittee(weights []primitives.MemberWeight) []interfaces.CommitteeMember {
	committee := make([]interfaces.CommitteeMember, len(weights))
	for i, weight := range weights {
		committee[i] = interfaces.CommitteeMember{
			Id:     []byte{byte(i)},
			Weight: weight,
		}
	}
	return committee
}

func TestIsQuorum(t *testing.T) {
	committee := genCommittee([]primitives.MemberWeight{1, 6, 10, 23, 60})

	ids := func(inds []int) []primitives.MemberId {
		_ids := make([]primitives.MemberId, len(inds))
		for i, ind := range inds {
			_ids[i] = committee[ind].Id
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
