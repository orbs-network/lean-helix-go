package leanhelixterm

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParticipating(t *testing.T) {
	myMemberId := primitives.MemberId("My ID")
	memberId1 := primitives.MemberId("Member 1")
	memberId2 := primitives.MemberId("Member 2")
	memberId3 := primitives.MemberId("Member 3")
	committeeMembers := []primitives.MemberId{myMemberId, memberId1, memberId2, memberId3}
	actual := isParticipatingInCommittee(myMemberId, committeeMembers)
	require.True(t, actual)
}

func TestParticipatingLastInList(t *testing.T) {
	myMemberId := primitives.MemberId("My ID")
	memberId1 := primitives.MemberId("Member 1")
	memberId2 := primitives.MemberId("Member 2")
	memberId3 := primitives.MemberId("Member 3")
	committeeMembers := []primitives.MemberId{memberId1, memberId2, memberId3, myMemberId}
	actual := isParticipatingInCommittee(myMemberId, committeeMembers)
	require.True(t, actual)
}

func TestNotParticipating(t *testing.T) {
	myMemberId := primitives.MemberId("My ID")
	memberId1 := primitives.MemberId("Member 1")
	memberId2 := primitives.MemberId("Member 2")
	memberId3 := primitives.MemberId("Member 3")
	committeeMembers := []primitives.MemberId{memberId1, memberId2, memberId3}
	actual := isParticipatingInCommittee(myMemberId, committeeMembers)
	require.False(t, actual)
}
