package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockproof"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
	"strings"
)

func CommitsToProof(log logger.LHLogger, blockHeight primitives.BlockHeight, myMemberId primitives.MemberId, keyManager interfaces.KeyManager, onCommit interfaces.OnCommitCallback) termincommittee.OnInCommitteeCommitCallback {
	return func(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage) {
		proof := blockproof.GenerateLeanHelixBlockProof(keyManager, commitMessages)
		committeeStr := commitMessagesToMemberIdsStr(commitMessages)
		log.Info(L.LC(blockHeight, math.MaxUint64, myMemberId), "Generated block proof with committee=%s", committeeStr)
		onCommit(ctx, block, proof.Raw())
	}
}

func commitMessagesToMemberIdsStr(messages []*interfaces.CommitMessage) string {
	ids := make([]string, 0)
	for _, cm := range messages {
		ids = append(ids, cm.Content().Sender().MemberId().String())
	}
	return strings.Join(ids, ",")
}
