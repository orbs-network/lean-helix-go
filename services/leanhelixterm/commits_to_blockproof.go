package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockproof"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
)

func CommitsToProof(keyManager interfaces.KeyManager, onCommit interfaces.OnCommitCallback) termincommittee.OnInCommitteeCommitCallback {
	return func(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage) {
		proof := blockproof.GenerateLeanHelixBlockProof(keyManager, commitMessages).Raw()
		onCommit(ctx, block, proof)
	}
}
