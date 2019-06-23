// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package blockproof

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

// assume commit messages are valid and still hold
func GenerateLeanHelixBlockProof(keyManager interfaces.KeyManager, commitMessages []*interfaces.CommitMessage) *protocol.BlockProof {
	blockHeight := commitMessages[0].BlockHeight()
	blockRefBuilder := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_COMMIT,
		InstanceId:  commitMessages[0].InstanceId(),
		BlockHeight: blockHeight,
		View:        commitMessages[0].View(),
		BlockHash:   commitMessages[0].Content().SignedHeader().BlockHash(),
	}

	cSendersBuilders := make([]*protocol.SenderSignatureBuilder, 0)
	cShares := make([]*protocol.SenderSignature, 0)
	for _, cm := range commitMessages {
		memberId := cm.Content().Sender().MemberId()
		cSendersBuilders = append(cSendersBuilders, &protocol.SenderSignatureBuilder{
			MemberId:  memberId,
			Signature: cm.Content().Sender().Signature(),
		})

		cShares = append(cShares, (&protocol.SenderSignatureBuilder{
			MemberId:  memberId,
			Signature: primitives.Signature(cm.Content().Share()),
		}).Build())
	}

	randomSeedSignature := keyManager.AggregateRandomSeed(blockHeight, cShares)
	return (&protocol.BlockProofBuilder{
		BlockRef:            blockRefBuilder,
		Nodes:               cSendersBuilders,
		RandomSeedSignature: randomSeedSignature,
	}).Build()
}
