// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type termMessagesHandlerMock struct {
	history []interfaces.ConsensusMessage
}

func NewTermMessagesHandlerMock() *termMessagesHandlerMock {
	return &termMessagesHandlerMock{}
}

func (t *termMessagesHandlerMock) HandleConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) error {
	t.history = append(t.history, message)

	return nil
}
