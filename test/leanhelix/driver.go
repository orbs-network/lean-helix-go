package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"math/rand"
)

type driver struct {
	instanceId          primitives.InstanceId
	communication       *mocks.CommunicationMock
	config              *interfaces.Config
	leadersByView       []primitives.MemberId
	mainLoop            *leanhelix.MainLoop
	electionTriggerMock *mocks.ElectionTriggerMock
}

func newDriver(logger interfaces.Logger, becomeLeaderInView byte, totalMembers byte) *driver {
	if becomeLeaderInView >= totalMembers {
		panic("current node must be in committee")
	}

	instanceId := primitives.InstanceId(0)
	discoveryMock := mocks.NewDiscovery()

	leadersByView := make([]primitives.MemberId, totalMembers)
	communications := make([]*mocks.CommunicationMock, totalMembers)

	for i := byte(0); i < totalMembers; i++ {
		leadersByView[i] = primitives.MemberId{i}
		communications[i] = mocks.NewCommunication(leadersByView[i], discoveryMock, logger)
		// set leadership rotation order through views
		discoveryMock.RegisterCommunication(leadersByView[i], communications[i])
	}

	currentMemberId := leadersByView[becomeLeaderInView]
	currentMemberCommunication := communications[becomeLeaderInView]

	membership := mocks.NewFakeMembership(currentMemberId, discoveryMock, false)

	keyManager := mocks.NewMockKeyManager(currentMemberId)
	keyManager.DisableConsensusMessageVerification()

	electionTriggerMock := mocks.NewMockElectionTrigger()
	config := mocks.NewMockConfig(
		logger,
		instanceId,
		membership,
		mocks.NewMockBlockUtils(currentMemberId, mocks.NewBlocksPool(nil), logger),
		keyManager,
		electionTriggerMock,
		currentMemberCommunication,
	)

	mainLoop := leanhelix.NewLeanHelix(config, nil, nil)

	return &driver{
		instanceId:          instanceId,
		communication:       currentMemberCommunication,
		config:              config,
		leadersByView:       leadersByView,
		mainLoop:            mainLoop,
		electionTriggerMock: electionTriggerMock,
	}
}

func (d *driver) handleViewChangeMessage(ctx context.Context, hv *state.HeightView, fromLeaderAtView byte) {
	randomSeed := rand.Uint64()
	messageFactory := messagesfactory.NewMessageFactory(d.instanceId, d.config.KeyManager, d.leadersByView[fromLeaderAtView], randomSeed)

	message := messageFactory.CreateViewChangeMessage(hv.Height(), hv.View(), nil)

	d.mainLoop.HandleConsensusMessage(ctx, message.ToConsensusRawMessage())
}
