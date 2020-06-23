package testhelpers

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

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
