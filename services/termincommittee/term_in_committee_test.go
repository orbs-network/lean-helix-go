package termincommittee

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStrFunc(t *testing.T) {
	var memberId primitives.MemberId
	memberId = []byte{16, 32, 48, 64, 80, 96}
	memberIdStr := Str(memberId)
	require.Equal(t, "102030", memberIdStr, "bad translation of memberId to string")

	require.Equal(t, "", Str(nil), "bad translation of memberId to string")

}

func TestIsLeader_CandidateIsLeaderForViewAndCommittee(t *testing.T) {
	committeeMembers := []primitives.MemberId{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, []byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19}, []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}, []byte{30, 31, 32, 33, 34, 35, 36, 37, 38, 39}}
	leaderCandidate := primitives.MemberId([]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19})

	err := isLeaderOfViewForThisCommittee(leaderCandidate, primitives.View(1), committeeMembers)
	require.Nil(t, err, "expected leader to be %s but got: %s", leaderCandidate, err)
}

func TestIsLeader_CandidateIsNotLeaderForView(t *testing.T) {
	committeeMembers := []primitives.MemberId{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, []byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19}, []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}, []byte{30, 31, 32, 33, 34, 35, 36, 37, 38, 39}}
	leaderCandidate := primitives.MemberId([]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19})

	err := isLeaderOfViewForThisCommittee(leaderCandidate, primitives.View(0), committeeMembers)
	t.Log(err)
	require.Error(t, err, "expected leader to be %s but got: %s", Str(leaderCandidate), err)
}
