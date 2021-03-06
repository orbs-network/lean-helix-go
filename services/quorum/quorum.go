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

func GetWeights(members []interfaces.CommitteeMember) []primitives.MemberWeight {
	weights := make([]primitives.MemberWeight, len(members))
	for i, member := range members {
		weights[i] = member.Weight
	}
	return weights
}

func CalcQuorumWeight(committeeWeights []primitives.MemberWeight) uint {
	sum := uint(0)
	for _, weight := range committeeWeights {
		sum += uint(weight)
	}

	if sum == 0 {
		return 1
	}

	return sum - calcF(sum)
}

func CalcByzMaxWeight(committeeWeights []primitives.MemberWeight) uint {
	sum := uint(0)
	for _, weight := range committeeWeights {
		sum += uint(weight)
	}

	if sum == 0 {
		return sum
	}

	return calcF(sum)
}

func calcF(totalWeight uint) uint {
	return uint(math.Floor(float64(totalWeight-1) / 3))
}

func IsQuorum(committeeSubset []primitives.MemberId, allCommitteeMembers []interfaces.CommitteeMember) (bool, uint, uint) {
	weight := getCommitteeSubsetWeight(committeeSubset, allCommitteeMembers)

	q := CalcQuorumWeight(GetWeights(allCommitteeMembers))
	return weight >= q, weight, q
}

func HasHonest(committeeSubset []primitives.MemberId, allCommitteeMembers []interfaces.CommitteeMember) (bool, uint, uint) {
	weight := getCommitteeSubsetWeight(committeeSubset, allCommitteeMembers)

	b := CalcByzMaxWeight(GetWeights(allCommitteeMembers))
	return weight > b, weight, b
}

func getCommitteeSubsetWeight(committeeSubset []primitives.MemberId, allCommitteeMembers []interfaces.CommitteeMember) uint {
	subsetIdsSet := make(map[string]bool)
	for _, id := range committeeSubset {
		subsetIdsSet[id.String()] = true
	}

	sum := uint(0)
	for _, member := range allCommitteeMembers {
		if subsetIdsSet[member.Id.String()] {
			sum += uint(member.Weight)
		}
	}
	return sum
}
