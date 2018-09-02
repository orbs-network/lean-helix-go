package proofsvalidator

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/go/test/builders"
	"github.com/orbs-network/lean-helix-go/go/test/keymanagermock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	keyManager := keymanagermock.NewMockKeyManager("Dummy PK")
	leaderKeyManager := keymanagermock.NewMockKeyManager("Leader PK")
	node1KeyManager := keymanagermock.NewMockKeyManager("Node 1")
	node2KeyManager := keymanagermock.NewMockKeyManager("Node 2")
	//node3KeyManager := keymanagermock.NewMockKeyManager("Node 3")

	const f = 1
	const term = 0
	const view = 0
	const targetTerm = term
	const targetView = view + 1
	block := builders.CreateBlock(builders.GenesisBlock)

	preprepareMessage := builders.CreatePrePrepareMessage(leaderKeyManager, term, view, block)
	prepareMessage1 := builders.CreatePrepareMessage(node1KeyManager, term, view, block)
	prepareMessage2 := builders.CreatePrepareMessage(node2KeyManager, term, view, block)
	preparedProof := &lh.PreparedProof{
		PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
		PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
	}

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: nil,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
			PrepareBlockRefMessages:   nil,
		}
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, nil, f, keyManager)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		keyManager := keymanagermock.NewMockKeyManager("Dummy PK", "Leader PK")
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		keyManager := keymanagermock.NewMockKeyManager("Dummy PK", "Node 2")
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1},
		}
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithTerm", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(666, targetView, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof with mismatching term")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetTerm, view, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView-1, preparedProof, f, keyManager)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager)
		require.True(t, result, "Did not approve a valid proof")
	})
}