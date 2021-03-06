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
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"strings"
)

func CommitsToProof(log logger.LHLogger, keyManager interfaces.KeyManager, onCommit interfaces.OnCommitCallback) termincommittee.OnInCommitteeCommitCallback {
	return func(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage) {
		proof := blockproof.GenerateLeanHelixBlockProof(keyManager, commitMessages)
		committeeStr := commitMessagesToCommitteeMemberIdsStr(commitMessages)
		height := block.Height()
		log.Debug("Generated block proof for H=%d with committee-size=%d, committee-members=%s", height, len(commitMessages), committeeStr)
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
