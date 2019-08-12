// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelixterm

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
)

type ConsensusMessagesFilter struct {
	handler    TermMessagesHandler
	keyManager interfaces.KeyManager
	randomSeed uint64
}

func NewConsensusMessagesFilter(handler TermMessagesHandler, keyManager interfaces.KeyManager, randomSeed uint64) *ConsensusMessagesFilter {
	return &ConsensusMessagesFilter{handler, keyManager, randomSeed}
}

func (mp *ConsensusMessagesFilter) HandleConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) error {
	if mp.handler == nil {
		return errors.New("mp.handler is nil")
	}

	switch message := message.(type) {
	case *interfaces.PreprepareMessage:
		mp.handler.HandlePrePrepare(ctx, message)

	case *interfaces.PrepareMessage:
		mp.handler.HandlePrepare(ctx, message)

	case *interfaces.CommitMessage:
		senderSignature := (&protocol.SenderSignatureBuilder{
			MemberId:  message.Content().Sender().MemberId(),
			Signature: primitives.Signature(message.Content().Share()),
		}).Build()

		randomSeedBytes := randomseed.RandomSeedToBytes(mp.randomSeed)
		if err := mp.keyManager.VerifyRandomSeed(message.BlockHeight(), randomSeedBytes, senderSignature); err != nil {
			return errors.Wrapf(err, "Failed in VerifyRandomSeed()")
		}
		mp.handler.HandleCommit(ctx, message)

	case *interfaces.ViewChangeMessage:
		mp.handler.HandleViewChange(ctx, message)

	case *interfaces.NewViewMessage:
		mp.handler.HandleNewView(ctx, message)

	default:
		panic(fmt.Sprintf("unknown message type: %T", message))
	}

	return nil
}
