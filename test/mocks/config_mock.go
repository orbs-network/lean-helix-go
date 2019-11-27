package mocks

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

func NewMockConfigSimple() *interfaces.Config {
	membership := NewFakeMembership([]byte{30, 30, 30}, nil, false)
	return NewMockConfig(
		nil,
		0,
		membership,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func NewMockConfig(logger interfaces.Logger, instanceId primitives.InstanceId, membership interfaces.Membership, blockUtils interfaces.BlockUtils, keyManager interfaces.KeyManager, electionSched interfaces.ElectionScheduler, communication interfaces.Communication, onNewView interfaces.OnNewViewCallback) *interfaces.Config {
	return &interfaces.Config{
		Logger:                  logger,
		Membership:              membership,
		BlockUtils:              blockUtils,
		KeyManager:              keyManager,
		OverrideElectionTrigger: electionSched,
		InstanceId:              instanceId,
		Communication:           communication,
		OnNewViewCB:             onNewView,
	}
}
