package test

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

// Unit tests for leanhelix_term

func TestReturnOkForMembersInCurrentHeight(t *testing.T) {
	// Instantiate, provide no members, and expect error
	pk := lh.PublicKey("PK")
	mockComm := builders.NewMockNetworkCommunication()
	mockBlockUtils := builders.NewMockBlockUtils(nil)
	mockElectionTrigger := builders.NewMockElectionTrigger()
	mockStorage := lh.NewInMemoryPBFTStorage()
	mockComm.When("GetMembersPKs", mock.Any).Return([]lh.PublicKey{pk})
	mockComm.When("SendPreprepare", mock.Any, mock.Any).Return()
	config := &lh.TermConfig{
		KeyManager:           builders.NewMockKeyManager(pk),
		NetworkCommunication: mockComm,
		Logger:               log.GetLogger(log.String("ID", "ID")),
		BlockUtils:           mockBlockUtils,
		ElectionTrigger:      mockElectionTrigger,
		Storage:              mockStorage,
	}
	term, _ := lh.NewLeanHelixTerm(config, 1, nil)
	require.NotNil(t, term, "should return new term if there are members in current height")
}

func TestReturnErrorIfNoMembersInCurrentHeight(t *testing.T) {
	// Instantiate, provide no members, and expect error

	pk := lh.PublicKey("PK")
	mockComm := builders.NewMockNetworkCommunication()
	mockBlockUtils := builders.NewMockBlockUtils(nil)
	mockElectionTrigger := builders.NewMockElectionTrigger()
	mockStorage := lh.NewInMemoryPBFTStorage()
	mockComm.When("GetMembersPKs", mock.Any).Return([]lh.PublicKey{})
	mockComm.When("SendPreprepare", mock.Any, mock.Any).Return()
	config := &lh.TermConfig{
		KeyManager:           builders.NewMockKeyManager(pk),
		NetworkCommunication: mockComm,
		Logger:               log.GetLogger(log.String("ID", "ID")),
		BlockUtils:           mockBlockUtils,
		ElectionTrigger:      mockElectionTrigger,
		Storage:              mockStorage,
	}
	_, err := lh.NewLeanHelixTerm(config, 1, nil)
	require.NotNil(t, err, "should return error if no members in current height")

}
