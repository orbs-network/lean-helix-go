package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type NewViewBuilder struct {
	leaderKeyManager lh.KeyManager
	votes            []*lh.ViewChangeMessageContentBuilder
	customPP         *lh.PreprepareContentBuilder
	block            lh.Block
	blockHeight      primitives.BlockHeight
	view             primitives.View
}

func (builder *NewViewBuilder) Build() *lh.NewViewMessage {
	messageFactory := lh.NewMessageFactory(builder.leaderKeyManager)

	ppmCB := builder.customPP
	if ppmCB == nil {
		ppmCB = messageFactory.CreatePreprepareMessageContentBuilder(builder.blockHeight, builder.view, builder.block, CalculateBlockHash(builder.block))
	}

	nvcb := messageFactory.CreateNewViewMessageContentBuilder(builder.blockHeight, builder.view, ppmCB, builder.votes)
	return lh.NewNewViewMessage(nvcb.Build(), builder.block)
}

func (builder *NewViewBuilder) LeadBy(keyManager lh.KeyManager) *NewViewBuilder {
	builder.leaderKeyManager = keyManager
	return builder
}

func (builder *NewViewBuilder) WithCustomPreprepare(keyManager lh.KeyManager, blockHeight primitives.BlockHeight, view primitives.View, block lh.Block) *NewViewBuilder {
	messageFactory := lh.NewMessageFactory(keyManager)
	builder.customPP = messageFactory.CreatePreprepareMessageContentBuilder(blockHeight, view, block, CalculateBlockHash(block))
	return builder
}

func (builder *NewViewBuilder) WithViewChangeVotes(votes []*lh.ViewChangeMessageContentBuilder) *NewViewBuilder {
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

func (builder *NewViewBuilder) OnBlock(block lh.Block) *NewViewBuilder {
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
	keyManager  lh.KeyManager
	blockHeight primitives.BlockHeight
	view        primitives.View
}

type VotesBuilder struct {
	voters []*Voter
}

func (builder *VotesBuilder) Build() []*lh.ViewChangeMessageContentBuilder {
	var votes []*lh.ViewChangeMessageContentBuilder
	for _, voter := range builder.voters {
		messageFactory := lh.NewMessageFactory(voter.keyManager)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(voter.blockHeight, voter.view, nil)
		votes = append(votes, vcmCB)
	}

	return votes
}

func (builder *VotesBuilder) WithVoter(keyManager lh.KeyManager, blockHeight primitives.BlockHeight, view primitives.View) *VotesBuilder {
	voter := &Voter{keyManager, blockHeight, view}
	builder.voters = append(builder.voters, voter)
	return builder
}

func NewVotesBuilder() *VotesBuilder {
	return &VotesBuilder{}
}

func ASimpleViewChangeVotes(membersKeyManagers []lh.KeyManager, blockHeight primitives.BlockHeight, view primitives.View) []*lh.ViewChangeMessageContentBuilder {
	builder := NewVotesBuilder()
	for _, keyManager := range membersKeyManagers {
		builder.WithVoter(keyManager, blockHeight, view)
	}
	return builder.Build()
}
