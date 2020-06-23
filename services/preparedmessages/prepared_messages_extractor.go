// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package preparedmessages

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type PreparedMessages struct {
	PreprepareMessage *interfaces.PreprepareMessage
	PrepareMessages   []*interfaces.PrepareMessage
}

// TODO are there missing verifications here? verify no repeated addresses
func ExtractPreparedMessages(blockHeight primitives.BlockHeight, latestPreparedView primitives.View, storage interfaces.Storage, isQuorum func([]primitives.MemberId) bool) *PreparedMessages {

	ppm, ok := storage.GetPreprepareFromView(blockHeight, latestPreparedView)
	if !ok {
		return nil
	}

	senderIds := storage.GetPrepareSendersIds(blockHeight, latestPreparedView, ppm.Content().SignedHeader().BlockHash())
	senderIds = append(senderIds, ppm.SenderMemberId())
	if !isQuorum(senderIds) {
		return nil
	}

	prepareMessages, ok := storage.GetPrepareMessages(blockHeight, latestPreparedView, ppm.Content().SignedHeader().BlockHash())
	if !ok {
		return nil
	}

	return &PreparedMessages{
		PreprepareMessage: ppm,
		PrepareMessages:   prepareMessages,
	}
}
