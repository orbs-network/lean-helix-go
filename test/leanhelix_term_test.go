package test

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

// Unit tests for leanhelix_term

func TestReturnErrorIfNoMembersInCurrentHeight(t *testing.T) {
	// Instantiate, provide no members, and expect error

	comm := builders.NewMockNetworkCommunication()
	comm.When("GetMembersPKs", mock.Any).Return(nil)
	config := &lh.TermConfig{
		NetworkCommunication: comm,
	}
	_, err := lh.NewLeanHelixTerm(config, 1, nil)
	require.NotNil(t, err, "should return error if no members in current height")

}
