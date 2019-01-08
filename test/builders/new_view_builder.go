package builders

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

type NewViewBuilder struct {
	instanceId       primitives.InstanceId
	leaderKeyManager interfaces.KeyManager
	leaderMemberId   primitives.MemberId
	votes            []*protocol.ViewChangeMessageContentBuilder
	customPP         *protocol.PreprepareContentBuilder
	block            interfaces.Block
	blockHeight      primitives.BlockHeight
	view             primitives.View
}

func (builder *NewViewBuilder) Build() *interfaces.NewViewMessage {
	messageFactory := messagesfactory.NewMessageFactory(builder.instanceId, builder.leaderKeyManager, builder.leaderMemberId, 0)

	ppmCB := builder.customPP
	if ppmCB == nil {
		ppmCB = messageFactory.CreatePreprepareMessageContentBuilder(builder.blockHeight, builder.view, builder.block, mocks.CalculateBlockHash(builder.block))
	}

	nvcb := messageFactory.CreateNewViewMessageContentBuilder(builder.blockHeight, builder.view, ppmCB, builder.votes)
	return interfaces.NewNewViewMessage(nvcb.Build(), builder.block)
}

func (builder *NewViewBuilder) LeadBy(keyManager interfaces.KeyManager, memberId primitives.MemberId) *NewViewBuilder {
	builder.leaderKeyManager = keyManager
	builder.leaderMemberId = memberId
	return builder
}

func (builder *NewViewBuilder) WithCustomPreprepare(instanceId primitives.InstanceId, keyManager interfaces.KeyManager, memberId primitives.MemberId, blockHeight primitives.BlockHeight, view primitives.View, block interfaces.Block) *NewViewBuilder {
	messageFactory := messagesfactory.NewMessageFactory(instanceId, keyManager, memberId, 0)
	builder.customPP = messageFactory.CreatePreprepareMessageContentBuilder(blockHeight, view, block, mocks.CalculateBlockHash(block))
	return builder
}

func (builder *NewViewBuilder) WithViewChangeVotes(votes []*protocol.ViewChangeMessageContentBuilder) *NewViewBuilder {
	builder.votes = votes
	return builder
}

func (builder *NewViewBuilder) OnBlockHeight(blockHeight primitives.BlockHeight) *NewViewBuilder {
	builder.blockHeight = blockHeight
	return builder
}

func (builder *NewViewBuilder) OnView(view primitives.View) *NewViewBuilder {
	builder.view = view
	return builder
}

func (builder *NewViewBuilder) OnBlock(block interfaces.Block) *NewViewBuilder {
	builder.block = block
	return builder
}

func NewNewViewBuilder() *NewViewBuilder {
	return &NewViewBuilder{
		customPP: nil,
	}
}

//////////////////// View Change Votes Builder //////////////////////////

type Vote struct {
	keyManager       interfaces.KeyManager
	memberId         primitives.MemberId
	blockHeight      primitives.BlockHeight
	view             primitives.View
	preparedMessages *preparedmessages.PreparedMessages
}

type VotesBuilder struct {
	instanceId primitives.InstanceId
	voters     []*Vote
}

func (builder *VotesBuilder) Build() []*protocol.ViewChangeMessageContentBuilder {
	var votes []*protocol.ViewChangeMessageContentBuilder
	for _, voter := range builder.voters {
		messageFactory := messagesfactory.NewMessageFactory(builder.instanceId, voter.keyManager, voter.memberId, 0)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(voter.blockHeight, voter.view, voter.preparedMessages)
		votes = append(votes, vcmCB)
	}

	return votes
}

func (builder *VotesBuilder) WithVote(keyManager interfaces.KeyManager, memberId primitives.MemberId, blockHeight primitives.BlockHeight, view primitives.View, preparedMessages *preparedmessages.PreparedMessages) *VotesBuilder {
	voter := &Vote{keyManager, memberId, blockHeight, view, preparedMessages}
	builder.voters = append(builder.voters, voter)
	return builder
}

func NewVotesBuilder(instanceId primitives.InstanceId) *VotesBuilder {
	return &VotesBuilder{instanceId: instanceId}
}

type Voter struct {
	KeyManager interfaces.KeyManager
	MemberId   primitives.MemberId
}

func ASimpleViewChangeVotes(instanceId primitives.InstanceId, voters []*Voter, blockHeight primitives.BlockHeight, view primitives.View) []*protocol.ViewChangeMessageContentBuilder {
	builder := NewVotesBuilder(instanceId)
	for _, voter := range voters {
		builder.WithVote(voter.KeyManager, voter.MemberId, blockHeight, view, nil)
	}
	return builder.Build()
}
