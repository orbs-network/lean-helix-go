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
	//keyManager := keymanagermock.NewMockKeyManager("Dummy PK", nil)
	leaderKeyManager := keymanagermock.NewMockKeyManager("Leader PK", nil)
	node1KeyManager := keymanagermock.NewMockKeyManager("Node 1", nil)
	node2KeyManager := keymanagermock.NewMockKeyManager("Node 2", nil)
	//node3KeyManager := keymanagermock.NewMockKeyManager("Node 3", nil)

	const term = 0
	const view = 0
	const targetTerm = term
	const targetView = view + 1
	block := builders.CreateBlock(builders.GenesisBlock)

	preprepareMessage := builders.CreatePrePrepareMessage(leaderKeyManager, term, view, block)
	preprepareBlockRefMessage := builders.BlockRefMessageFromPrePrepare(preprepareMessage)
	prepareMessage1 := builders.CreatePrepareMessage(node1KeyManager, term, view, block)
	prepareMessage2 := builders.CreatePrepareMessage(node2KeyManager, term, view, block)

	t.Run("TestProofsValidatorWithNoPrePrepare", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: nil,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareBlockRefMessage,
			PrepareBlockRefMessages:   nil,
		}
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, nil)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareBlockRefMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := proofsvalidator.ValidatePreparedProof(targetTerm, targetView, preparedProof)
		require.True(t, result, "Did not approve a valid proof")
	})
}
