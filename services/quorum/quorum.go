// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package quorum

import "math"

func CalcQuorumWeight(committeeWeights []uint) uint {
	sum := uint(0)
	for i := 0; i < len(committeeWeights); i++ {
		sum += committeeWeights[i]
	}

	if sum == 0 {
		return 1
	}

	return sum - uint(math.Floor(float64(sum-1)/3)) // TODO make this accurate!

	//f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	//return committeeMembersCount - f
}
