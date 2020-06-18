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

	prepareMessages, ok := storage.GetPrepareMessages(blockHeight, latestPreparedView, ppm.Content().SignedHeader().BlockHash())
	if !ok {
		return nil
	}

	senderIds := make([]primitives.MemberId, len(prepareMessages)+1)
	senderIds[0] = ppm.SenderMemberId()
	for i := 1; i <= len(prepareMessages); i++ {
		senderIds[i] = prepareMessages[i-1].SenderMemberId()
	}

	//if len(preparedMessages) < tic.QuorumSize-1 {
	if !isQuorum(senderIds) { // todo -1?
		return nil
	}

	return &PreparedMessages{
		PreprepareMessage: ppm,
		PrepareMessages:   prepareMessages,
	}
}
