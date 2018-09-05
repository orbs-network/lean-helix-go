package proofsvalidator

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	pv "github.com/orbs-network/lean-helix-go/go/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/go/test/builders"
	"github.com/orbs-network/lean-helix-go/go/test/inmemoryblockchain"
	"github.com/orbs-network/lean-helix-go/go/test/keymanagermock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProofsValidator(t *testing.T) {
	keyManager := keymanagermock.NewMockKeyManager("Dummy PK")
	leaderKeyManager := keymanagermock.NewMockKeyManager("Leader PK")
	node1KeyManager := keymanagermock.NewMockKeyManager("Node 1")
	node2KeyManager := keymanagermock.NewMockKeyManager("Node 2")

	membersPKs := []lh.PublicKey{"Leader PK", "Node 1", "Node 2", "Node 3"}
	calcLeaderPk := func(view lh.ViewCounter) lh.PublicKey {
		return membersPKs[view]
	}

	const f = 1
	const term = 0
	const view = 0
	const targetTerm = term
	const targetView = view + 1
	block := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	leaderMsgFactory := lh.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
	node1MsgFactory := lh.NewMessageFactory(builders.CalculateBlockHash, node1KeyManager)
	node2MsgFactory := lh.NewMessageFactory(builders.CalculateBlockHash, node2KeyManager)

	preprepareMessage := leaderMsgFactory.CreatePreprepareMessage(term, view, block)
	prepareMessage1 := node1MsgFactory.CreatePrepareMessage(term, view, block)
	prepareMessage2 := node2MsgFactory.CreatePrepareMessage(term, view, block)
	preparedProof := &lh.PreparedProof{
		PreprepareBlockRefMessage: preprepareMessage,
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
			PreprepareBlockRefMessage: preprepareMessage,
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
			PreprepareBlockRefMessage: preprepareMessage,
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
		mf := lh.NewMessageFactory(builders.CalculateBlockHash, noneMemberKeyManager)
		prepareMessage1 := mf.CreatePrepareMessage(term, view, block)
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a none member")
	})

	t.Run("TestProofsValidatorWithPrepareFromTheLeader", func(t *testing.T) {
		mf := lh.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
		prepareMessage1 := mf.CreatePrepareMessage(term, view, block)
		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: preprepareMessage,
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
			PreprepareBlockRefMessage: preprepareMessage,
			PrepareBlockRefMessages:   []*lh.PrepareMessage{prepareMessage1, prepareMessage2},
		}
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with a mismatching view to leader")
	})

	t.Run("TestProofsValidatorWithMismatchingContent", func(t *testing.T) {
		// Good proof //
		const term = 5
		const view = 0
		const targetTerm = term
		const targetView = view + 1

		leaderMF := lh.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
		node1MF := lh.NewMessageFactory(builders.CalculateBlockHash, node1KeyManager)
		node2MF := lh.NewMessageFactory(builders.CalculateBlockHash, node2KeyManager)

		// TODO Maybe can use node1MsgFactory instead of creating node1MF here (same for leader and node2)
		// Good proof //
		goodPrepareProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(term, view, block),
			},
		}
		actualGood := pv.ValidatePreparedProof(targetTerm, targetView, goodPrepareProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, actualGood, "Did not approve a valid proof")

		// Mismatching term //
		badTermProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(666, view, block),
			},
		}
		actualBadTerm := pv.ValidatePreparedProof(targetTerm, targetView, badTermProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadTerm, "Did not reject mismatching term")

		// Mismatching view //
		badViewProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(term, 666, block),
			},
		}
		actualBadView := pv.ValidatePreparedProof(targetTerm, targetView, badViewProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadView, "Did not reject mismatching view")

		// Mismatching blockHash //
		otherBlock := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
		badBlockHashProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node2MF.CreatePrepareMessage(term, view, otherBlock),
			},
		}
		actualBadBlockHash := pv.ValidatePreparedProof(targetTerm, targetView, badBlockHashProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, actualBadBlockHash, "Did not reject mismatching block hash")
	})

	t.Run("TestProofsValidatorWithDuplicate prepare sender PK", func(t *testing.T) {
		leaderMF := lh.NewMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
		node1MF := lh.NewMessageFactory(builders.CalculateBlockHash, node1KeyManager)

		preparedProof := &lh.PreparedProof{
			PreprepareBlockRefMessage: leaderMF.CreatePreprepareMessage(term, view, block),
			PrepareBlockRefMessages: []*lh.PrepareMessage{
				node1MF.CreatePrepareMessage(term, view, block),
				node1MF.CreatePrepareMessage(term, view, block),
			},
		}

		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.False(t, result, "Did not reject a proof with duplicate sender PK")
	})

	t.Run("TestProofsValidatorWithNoProof", func(t *testing.T) {
		result := pv.ValidatePreparedProof(targetTerm, targetView, preparedProof, f, keyManager, &membersPKs, calcLeaderPk)
		require.True(t, result, "Did not approve a valid proof")
	})
}
