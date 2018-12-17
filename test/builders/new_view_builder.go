package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type NewViewBuilder struct {
	leaderKeyManager leanhelix.KeyManager
	votes            []*protocol.ViewChangeMessageContentBuilder
	customPP         *protocol.PreprepareContentBuilder
	block            leanhelix.Block
	blockHeight      primitives.BlockHeight
	view             primitives.View
}

func (builder *NewViewBuilder) Build() *leanhelix.NewViewMessage {
	messageFactory := leanhelix.NewMessageFactory(builder.leaderKeyManager)

	ppmCB := builder.customPP
	if ppmCB == nil {
		ppmCB = messageFactory.CreatePreprepareMessageContentBuilder(builder.blockHeight, builder.view, builder.block, CalculateBlockHash(builder.block))
	}

	nvcb := messageFactory.CreateNewViewMessageContentBuilder(builder.blockHeight, builder.view, ppmCB, builder.votes)
	return leanhelix.NewNewViewMessage(nvcb.Build(), builder.block)
}

func (builder *NewViewBuilder) LeadBy(keyManager leanhelix.KeyManager) *NewViewBuilder {
	builder.leaderKeyManager = keyManager
	return builder
}

func (builder *NewViewBuilder) WithCustomPreprepare(keyManager leanhelix.KeyManager, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) *NewViewBuilder {
	messageFactory := leanhelix.NewMessageFactory(keyManager)
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

type Voter struct {
	keyManager       leanhelix.KeyManager
	blockHeight      primitives.BlockHeight
	view             primitives.View
	preparedMessages *leanhelix.PreparedMessages
}

type VotesBuilder struct {
	voters []*Voter
}

func (builder *VotesBuilder) Build() []*protocol.ViewChangeMessageContentBuilder {
	var votes []*protocol.ViewChangeMessageContentBuilder
	for _, voter := range builder.voters {
		messageFactory := leanhelix.NewMessageFactory(voter.keyManager)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(voter.blockHeight, voter.view, voter.preparedMessages)
		votes = append(votes, vcmCB)
	}

	return votes
}

func (builder *VotesBuilder) WithVoter(keyManager leanhelix.KeyManager, blockHeight primitives.BlockHeight, view primitives.View, preparedMessages *leanhelix.PreparedMessages) *VotesBuilder {
	voter := &Voter{keyManager, blockHeight, view, preparedMessages}
	builder.voters = append(builder.voters, voter)
	return builder
}

func NewVotesBuilder() *VotesBuilder {
	return &VotesBuilder{}
}

func ASimpleViewChangeVotes(membersKeyManagers []leanhelix.KeyManager, blockHeight primitives.BlockHeight, view primitives.View) []*protocol.ViewChangeMessageContentBuilder {
	builder := NewVotesBuilder()
	for _, keyManager := range membersKeyManagers {
		builder.WithVoter(keyManager, blockHeight, view, nil)
	}
	return builder.Build()
}
