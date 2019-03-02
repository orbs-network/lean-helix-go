package termincommittee

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsParticipatingInCommittee(t *testing.T) {
	var MY_MEMBER_ID = []byte{0x01}
	var OTHER_MEMBER_A = []byte{0x02}
	var OTHER_MEMBER_B = []byte{0x03}
	var OTHER_MEMBER_C = []byte{0x04}

	tests := []struct {
		name                  string
		committee             []primitives.MemberId
		expectedParticipating bool
		expectedOthers        []primitives.MemberId
	}{
		{
			name:                  "Empty",
			committee:             []primitives.MemberId{},
			expectedParticipating: false,
			expectedOthers:        []primitives.MemberId{},
		},
		{
			name:                  "JustMe",
			committee:             []primitives.MemberId{MY_MEMBER_ID},
			expectedParticipating: true,
			expectedOthers:        []primitives.MemberId{},
		},
		{
			name:                  "EverybodyElse",
			committee:             []primitives.MemberId{OTHER_MEMBER_A, OTHER_MEMBER_B, OTHER_MEMBER_C},
			expectedParticipating: false,
			expectedOthers:        []primitives.MemberId{OTHER_MEMBER_A, OTHER_MEMBER_B, OTHER_MEMBER_C},
		},
		{
			name:                  "Everybody",
			committee:             []primitives.MemberId{MY_MEMBER_ID, OTHER_MEMBER_A, OTHER_MEMBER_B, OTHER_MEMBER_C},
			expectedParticipating: true,
			expectedOthers:        []primitives.MemberId{OTHER_MEMBER_A, OTHER_MEMBER_B, OTHER_MEMBER_C},
		},
		{
			name:                  "WithoutSomebody",
			committee:             []primitives.MemberId{MY_MEMBER_ID, OTHER_MEMBER_B, OTHER_MEMBER_C},
			expectedParticipating: true,
			expectedOthers:        []primitives.MemberId{OTHER_MEMBER_B, OTHER_MEMBER_C},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			participating, others := isParticipatingInCommittee(MY_MEMBER_ID, tt.committee)
			require.Equal(t, tt.expectedParticipating, participating, "participating should match")
			require.ElementsMatch(t, tt.expectedOthers, others, "others should match")

		})
	}
}
