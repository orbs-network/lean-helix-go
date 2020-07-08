package testhelpers

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

func EvenWeights(n int) []primitives.MemberWeight {
	weights := make([]primitives.MemberWeight, n)
	for i := 0; i < n; i++ {
		weights[i] = 1
	}
	return weights
}

func GenMembers(ids []primitives.MemberId) []interfaces.CommitteeMember {
	members := make([]interfaces.CommitteeMember, len(ids))
	for i := 0; i < len(ids); i++ {
		members[i] = interfaces.CommitteeMember{
			Id:     ids[i],
			Weight: 1,
		}
	}
	return members
}

func GenMembersWithWeights(ids []primitives.MemberId, weights []primitives.MemberWeight) []interfaces.CommitteeMember {
	members := make([]interfaces.CommitteeMember, len(ids))
	for i := 0; i < len(ids); i++ {
		members[i] = interfaces.CommitteeMember{
			Id:     ids[i],
			Weight: weights[i],
		}
	}
	return members
}
