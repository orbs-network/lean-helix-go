// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

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
		committeeStr := commitMessagesToCommitteeMemberIdsStr(commitMessages)
		log.Info(L.LC(blockHeight, math.MaxUint64, myMemberId), "Generated block proof with committee-size=%d, committee-members=%s", len(commitMessages), committeeStr)
		onCommit(ctx, block, proof.Raw())
	}
}

func commitMessagesToCommitteeMemberIdsStr(messages []*interfaces.CommitMessage) string {
	committeeMemberIds := make([]string, 0)
	for _, cm := range messages {
		committeeMemberId := cm.Content().Sender().MemberId().String()
		committeeMemberIds = append(committeeMemberIds, committeeMemberId)
	}
	return strings.Join(committeeMemberIds, ",")
}
