package test

import (
	"context"
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

// Unit tests for leanhelix_term

func TestReturnOkForMembersInCurrentHeight(t *testing.T) {

	t.Skip()
	// Instantiate, provide no members, and expect error
	ctx := context.Background()
	pk := Ed25519PublicKey("PK")
	mockComm := builders.NewMockNetworkCommunication()
	mockBlockUtils := builders.NewMockBlockUtils(nil)
	mockElectionTrigger := builders.NewMockElectionTrigger()
	mockStorage := builders.NewMockStorage()
	mockComm.When("RequestOrderedCommittee", mock.Any).Return([]Ed25519PublicKey{pk})
	mockComm.When("SendMessage", mock.Any, mock.Any, mock.Any).Return()
	mockStorage.When("StorePreprepare", mock.Any).Return(true)
	mockKeyManager := builders.NewMockKeyManager(pk)
	config := &lh.TermConfig{
		KeyManager:           mockKeyManager,
		NetworkCommunication: mockComm,
		Logger:               log.GetLogger(log.String("ID", "ID")),
		BlockUtils:           mockBlockUtils,
		ElectionTrigger:      mockElectionTrigger,
		Storage:              mockStorage,
	}
	term, _ := lh.NewLeanHelixTerm(ctx, config, 1, nil)
	require.NotNil(t, term, "should return new term if there are members in current height")
}

func TestReturnErrorIfNoMembersInCurrentHeight(t *testing.T) {
	// Instantiate, provide no members, and expect error
	ctx := context.Background()
	pk := Ed25519PublicKey("PK")
	mockComm := builders.NewMockNetworkCommunication()
	mockBlockUtils := builders.NewMockBlockUtils(nil)
	mockElectionTrigger := builders.NewMockElectionTrigger()
	mockStorage := lh.NewInMemoryStorage()
	mockComm.When("RequestOrderedCommittee", mock.Any).Return([]Ed25519PublicKey{})
	mockComm.When("SendMessage", mock.Any, mock.Any, mock.Any).Return()
	config := &lh.TermConfig{
		KeyManager:           builders.NewMockKeyManager(pk),
		NetworkCommunication: mockComm,
		Logger:               log.GetLogger(log.String("ID", "ID")),
		BlockUtils:           mockBlockUtils,
		ElectionTrigger:      mockElectionTrigger,
		Storage:              mockStorage,
	}
	_, err := lh.NewLeanHelixTerm(ctx, config, 1, nil)
	require.NotNil(t, err, "should return error if no members in current height")

}
