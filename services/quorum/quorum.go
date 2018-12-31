package quorum

import "math"

func CalcQuorumSize(committeeMembersCount int) int {
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}
