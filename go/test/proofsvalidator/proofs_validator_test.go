package proofsvalidator

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	pv "github.com/orbs-network/lean-helix-go/go/proofsvalidator"
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
	membersPKs := []lh.PublicKey{"Leader PK", "Node 1", "Node 2", "Node 3"}
	calcLeaderPk := func(view lh.ViewCounter) lh.PublicKey {
		return membersPKs[view]
	}

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
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have a preprepare message")
	})

	t.Run("TestProofsValidatorWithNoPrepares", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
			PrepareBlockRefMessages:   nil,
		}
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not have prepare messages")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := pv.ValidatePreparedProof(targetTerm, targetView, nil, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a nil proof")
	})

	t.Run("TestProofsValidatorWithBadPreprepareSignature", func(t *testing.T) {
		keyManager := keymanagermock.NewMockKeyManager("Dummy PK", "Leader PK")
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass preprepare signature validation")
	})

	t.Run("TestProofsValidatorWithBadPrepareSignature", func(t *testing.T) {
		keyManager := keymanagermock.NewMockKeyManager("Dummy PK", "Node 2")
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof that did not pass prepare signature validation")
	})

	t.Run("TestProofsValidatorWithNotEnoughPrepareMessages", func(t *testing.T) {
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1},
		}
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with not enough prepares")
	})

	t.Run("TestProofsValidatorWithTerm", func(t *testing.T) {
		result := pv.ValidatePreparedProof(666, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with mismatching term")
	})

	t.Run("TestProofsValidatorWithTheSameView", func(t *testing.T) {
		result := pv.ValidatePreparedProof(targetTerm, view, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with equal targetView")
	})

	t.Run("TestProofsValidatorWithTheSmallerView", func(t *testing.T) {
		result := pv.ValidatePreparedProof(targetTerm, targetView-1, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with smaller targetView")
	})

	t.Run("TestProofsValidatorWithANoneMember", func(t *testing.T) {
		noneMemberKeyManager := keymanagermock.NewMockKeyManager("Not in members PK")
		prepareMessage1 := builders.CreatePrepareMessage(noneMemberKeyManager, term, view, block)
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		prepareMessage1 := builders.CreatePrepareMessage(leaderKeyManager, term, view, block)
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a prepare from the leader")
	})

	t.Run("TestProofsValidatorWithMismatchingViewToLeader", func(t *testing.T) {
		calcLeaderPk := func(view lh.ViewCounter) lh.PublicKey {
			return "Some other node PK"
		}
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage.BlockRefMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")
	})

	t.Run("TestProofsValidatorWithMismatchingView", func(t *testing.T) {
		// Good proof //
		const term = 5
		const view = 0
		const targetTerm = term
		const targetView = view + 1

		// Good proof //
		goodPrepareProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: builders.CreatePrePrepareMessage(leaderKeyManager, term, view, block).BlockRefMessage,
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				builders.CreatePrepareMessage(node1KeyManager, term, view, block),
				builders.CreatePrepareMessage(node1KeyManager, term, view, block),
			},
		}
		actualGood := pv.ValidatePreparedProof(targetTerm, targetView, goodPrepareProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching term //
		badTermProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: builders.CreatePrePrepareMessage(leaderKeyManager, term, view, block).BlockRefMessage,
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				builders.CreatePrepareMessage(node1KeyManager, term, view, block),
				builders.CreatePrepareMessage(node1KeyManager, 666, view, block),
			},
		}
		actualBadTerm := pv.ValidatePreparedProof(targetTerm, targetView, badTermProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadTerm, "Did not reject mismatching term")

		// Mismatching view //
		badViewProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: builders.CreatePrePrepareMessage(leaderKeyManager, term, view, block).BlockRefMessage,
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				builders.CreatePrepareMessage(node1KeyManager, term, view, block),
				builders.CreatePrepareMessage(node1KeyManager, term, 666, block),
			},
		}
		actualBadView := pv.ValidatePreparedProof(targetTerm, targetView, badViewProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := builders.CreateBlock(builders.GenesisBlock)
		badBlockHashProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: builders.CreatePrePrepareMessage(leaderKeyManager, term, view, block).BlockRefMessage,
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				builders.CreatePrepareMessage(node1KeyManager, term, view, block),
				builders.CreatePrepareMessage(node1KeyManager, term, view, otherBlock),
			},
		}
		actualBadBlockHash := pv.ValidatePreparedProof(targetTerm, targetView, badBlockHashProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a valid proof")
	})
}
