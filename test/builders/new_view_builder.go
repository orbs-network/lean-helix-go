package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type NewViewBuilder struct {
	leaderKeyManager leanhelix.KeyManager
	leaderMemberId   primitives.MemberId
	votes            []*protocol.ViewChangeMessageContentBuilder
	customPP         *protocol.PreprepareContentBuilder
	block            leanhelix.Block
	blockHeight      primitives.BlockHeight
	view             primitives.View
}

func (builder *NewViewBuilder) Build() *leanhelix.NewViewMessage {
	messageFactory := leanhelix.NewMessageFactory(builder.leaderKeyManager, builder.leaderMemberId)

	ppmCB := builder.customPP
	if ppmCB == nil {
		ppmCB = messageFactory.CreatePreprepareMessageContentBuilder(builder.blockHeight, builder.view, builder.block, CalculateBlockHash(builder.block))
	}

	nvcb := messageFactory.CreateNewViewMessageContentBuilder(builder.blockHeight, builder.view, ppmCB, builder.votes)
	return leanhelix.NewNewViewMessage(nvcb.Build(), builder.block)
}

func (builder *NewViewBuilder) LeadBy(keyManager leanhelix.KeyManager, memberId primitives.MemberId) *NewViewBuilder {
	builder.leaderKeyManager = keyManager
	builder.leaderMemberId = memberId
	return builder
}

func (builder *NewViewBuilder) WithCustomPreprepare(keyManager leanhelix.KeyManager, memberId primitives.MemberId, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) *NewViewBuilder {
	messageFactory := leanhelix.NewMessageFactory(keyManager, memberId)
	builder.customPP = messageFactory.CreatePreprepareMessageContentBuilder(blockHeight, view, block, CalculateBlockHash(block))
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

func (builder *NewViewBuilder) OnBlock(block leanhelix.Block) *NewViewBuilder {
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
	keyManager       leanhelix.KeyManager
	memberId         primitives.MemberId
	blockHeight      primitives.BlockHeight
	view             primitives.View
	preparedMessages *leanhelix.PreparedMessages
}

type VotesBuilder struct {
	voters []*Vote
}

func (builder *VotesBuilder) Build() []*protocol.ViewChangeMessageContentBuilder {
	var votes []*protocol.ViewChangeMessageContentBuilder
	for _, voter := range builder.voters {
		messageFactory := leanhelix.NewMessageFactory(voter.keyManager, voter.memberId)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(voter.blockHeight, voter.view, voter.preparedMessages)
		votes = append(votes, vcmCB)
	}

	return votes
}

func (builder *VotesBuilder) WithVote(keyManager leanhelix.KeyManager, memberId primitives.MemberId, blockHeight primitives.BlockHeight, view primitives.View, preparedMessages *leanhelix.PreparedMessages) *VotesBuilder {
	voter := &Vote{keyManager, memberId, blockHeight, view, preparedMessages}
	builder.voters = append(builder.voters, voter)
	return builder
}

func NewVotesBuilder() *VotesBuilder {
	return &VotesBuilder{}
}

type Voter struct {
	KeyManager leanhelix.KeyManager
	MemberId   primitives.MemberId
}

func ASimpleViewChangeVotes(voters []*Voter, blockHeight primitives.BlockHeight, view primitives.View) []*protocol.ViewChangeMessageContentBuilder {
	builder := NewVotesBuilder()
	for _, voter := range voters {
		builder.WithVote(voter.KeyManager, voter.MemberId, blockHeight, view, nil)
	}
	return builder.Build()
}
