// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package quorum

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
)

func GetWeights(members []interfaces.CommitteeMember) []uint64 {
	weights := make([]uint64, len(members))
	for i := 0; i < len(members); i++ {
		weights[i] = members[i].Weight
	}
	return weights
}

func CalcQuorumWeight(committeeWeights []uint64) uint {
	sum := uint(0)
	for i := 0; i < len(committeeWeights); i++ {
		sum += uint(committeeWeights[i])
	}

	if sum == 0 {
		return 1
	}

	return sum - uint(math.Floor(float64(sum-1)/3)) // TODO make this accurate!

	//f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	//return committeeMembersCount - f
}

func IsQuorum(committeeSubset []primitives.MemberId, allCommitteeMembers []interfaces.CommitteeMember) (bool, uint, uint) {
	subsetIdsSet := make(map[string]bool)
	for i := 0; i < len(committeeSubset); i++ {
		subsetIdsSet[committeeSubset[i].String()] = true
	}

	sum := uint(0)
	for i := 0; i < len(allCommitteeMembers); i++ {
		if subsetIdsSet[allCommitteeMembers[i].Id.String()] {
			sum += uint(allCommitteeMembers[i].Weight)
		}
	}

	q := CalcQuorumWeight(GetWeights(allCommitteeMembers))
	return sum >= q, sum, q
}
