package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

type driver struct {
	instanceId          primitives.InstanceId
	communication       *mocks.CommunicationMock
	config              *interfaces.Config
	leadersByView       []primitives.MemberId
	mainLoop            *leanhelix.MainLoop
	electionTriggerMock *mocks.ElectionTriggerMock
	discovery           *mocks.Discovery
}

func newDriver(logger interfaces.Logger, becomeLeaderInView byte, totalMembers byte, onCommitCallback interfaces.OnCommitCallback) *driver {
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

	mainLoop := leanhelix.NewLeanHelix(config, onCommitCallback, nil)

	return &driver{
		instanceId:          instanceId,
		communication:       currentMemberCommunication,
		config:              config,
		leadersByView:       leadersByView,
		mainLoop:            mainLoop,
		electionTriggerMock: electionTriggerMock,
		discovery:           discoveryMock,
	}
}

func (d *driver) start(ctx context.Context, t *testing.T) {
	d.mainLoop.Run(ctx)
	err := d.mainLoop.UpdateState(ctx, interfaces.GenesisBlock, nil)
	require.NoError(t, err)

	require.True(t, test.Eventually(time.Second, func() bool {
		return d.electionTriggerMock.GetRegisteredHeight() == 1
	}))
}

func (d *driver) handleViewChangeMessage(ctx context.Context, hv *state.HeightView, fromLeaderAtView byte) {
	randomSeed := rand.Uint64()
	messageFactory := messagesfactory.NewMessageFactory(d.instanceId, d.config.KeyManager, d.leadersByView[fromLeaderAtView], randomSeed)

	message := messageFactory.CreateViewChangeMessage(hv.Height(), hv.View(), nil)

	d.mainLoop.HandleConsensusMessage(ctx, message.ToConsensusRawMessage())
}

func (d *driver) waitForSentPreprepareMessage(t *testing.T, i int) *interfaces.PreprepareMessage {
	require.True(t, test.Eventually(time.Second, func() bool {
		return len(d.communication.GetSentMessages(protocol.LEAN_HELIX_PREPREPARE)) >= i
	}), "expected a preprepare message to be sent within timeout")
	message := interfaces.ToConsensusMessage(d.communication.GetSentMessages(protocol.LEAN_HELIX_PREPREPARE)[i-1]).(*interfaces.PreprepareMessage)
	return message
}

func (d *driver) waitForSentCommitMessage(t *testing.T, i int) *interfaces.CommitMessage {
	require.True(t, test.Eventually(time.Second, func() bool {
		return len(d.communication.GetSentMessages(protocol.LEAN_HELIX_COMMIT)) >= i
	}), "expected a commit message to be sent within timeout")
	message := interfaces.ToConsensusMessage(d.communication.GetSentMessages(protocol.LEAN_HELIX_COMMIT)[i-1]).(*interfaces.CommitMessage)
	return message
}

func (d *driver) handlePrepareMessage(ctx context.Context, from primitives.MemberId, height primitives.BlockHeight, view primitives.View, block interfaces.Block) {
	message := builders.APrepareMessage(d.instanceId, mocks.NewMockKeyManager(from), from, height, view, block)
	d.mainLoop.HandleConsensusMessage(ctx, message.ToConsensusRawMessage())
}

func (d *driver) handleCommitMessage(ctx context.Context, from primitives.MemberId, height primitives.BlockHeight, view primitives.View, block interfaces.Block, randomSeed uint64) {
	message := builders.ACommitMessage(d.instanceId, mocks.NewMockKeyManager(from), from, height, view, block, randomSeed)
	d.mainLoop.HandleConsensusMessage(ctx, message.ToConsensusRawMessage())
}

func (d *driver) handlePreprepareMessage(ctx context.Context, from primitives.MemberId, height primitives.BlockHeight, view primitives.View, block interfaces.Block, randomSeed uint64) {
	message := builders.APreprepareMessage(d.instanceId, mocks.NewMockKeyManager(from), from, height, view, block)
	d.mainLoop.HandleConsensusMessage(ctx, message.ToConsensusRawMessage())
}
