package proofsvalidator

import lh "github.com/orbs-network/lean-helix-go/go/leanhelix"

func ValidatePreparedProof(
	targetTerm lh.BlockHeight,
	targetView lh.ViewCounter,
	preparedProof *lh.PreparedProof) bool {
	if preparedProof == nil {
		return true
	}

	if preparedProof.PreprepareBlockRefMessage == nil {
		return false
	}

	if preparedProof.PrepareBlockRefMessages == nil {
		return false
	}

	return true
}
