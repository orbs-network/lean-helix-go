// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelixterm

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/testhelpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParticipating(t *testing.T) {
	myMemberId := primitives.MemberId("My ID")
	memberId1 := primitives.MemberId("Member 1")
	memberId2 := primitives.MemberId("Member 2")
	memberId3 := primitives.MemberId("Member 3")
	committeeMembers := testhelpers.GenMembers([]primitives.MemberId{myMemberId, memberId1, memberId2, memberId3})
	actual := isParticipatingInTerm(myMemberId, committeeMembers)
	require.True(t, actual)
}

func TestParticipatingLastInList(t *testing.T) {
	myMemberId := primitives.MemberId("My ID")
	memberId1 := primitives.MemberId("Member 1")
	memberId2 := primitives.MemberId("Member 2")
	memberId3 := primitives.MemberId("Member 3")
	committeeMembers := testhelpers.GenMembers([]primitives.MemberId{memberId1, memberId2, memberId3, myMemberId})
	actual := isParticipatingInTerm(myMemberId, committeeMembers)
	require.True(t, actual)
}

func TestNotParticipating(t *testing.T) {
	myMemberId := primitives.MemberId("My ID")
	memberId1 := primitives.MemberId("Member 1")
	memberId2 := primitives.MemberId("Member 2")
	memberId3 := primitives.MemberId("Member 3")
	committeeMembers := testhelpers.GenMembers([]primitives.MemberId{memberId1, memberId2, memberId3})
	actual := isParticipatingInTerm(myMemberId, committeeMembers)
	require.False(t, actual)
}
