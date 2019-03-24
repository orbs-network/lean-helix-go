// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package builders

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Sender interface {
	GetKeyManager() interfaces.KeyManager
	GetMemberId() primitives.MemberId
}

func CreatePreparedMessages(
	instanceId primitives.InstanceId,
	leader Sender,
	members []Sender,
	blockHeight primitives.BlockHeight,
	view primitives.View,
	block interfaces.Block) *preparedmessages.PreparedMessages {

	PPMessage := APreprepareMessage(instanceId, leader.GetKeyManager(), leader.GetMemberId(), blockHeight, view, block)
	PMessages := make([]*interfaces.PrepareMessage, len(members))

	for i, member := range members {
		PMessages[i] = APrepareMessage(instanceId, member.GetKeyManager(), member.GetMemberId(), blockHeight, view, block)
	}

	return &preparedmessages.PreparedMessages{
		PreprepareMessage: PPMessage,
		PrepareMessages:   PMessages,
	}
}
