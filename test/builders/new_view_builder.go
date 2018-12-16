package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type NewViewBuilder struct {
	leaderKeyManager lh.KeyManager
	votes            []*protocol.ViewChangeMessageContentBuilder
	customPP         *protocol.PreprepareContentBuilder
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
	keyManager       lh.KeyManager
	blockHeight      primitives.BlockHeight
	view             primitives.View
	preparedMessages *lh.PreparedMessages
}

type VotesBuilder struct {
	voters []*Voter
}

func (builder *VotesBuilder) Build() []*protocol.ViewChangeMessageContentBuilder {
	var votes []*protocol.ViewChangeMessageContentBuilder
	for _, voter := range builder.voters {
		messageFactory := lh.NewMessageFactory(voter.keyManager)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(voter.blockHeight, voter.view, voter.preparedMessages)
		votes = append(votes, vcmCB)
	}

	return votes
}

func (builder *VotesBuilder) WithVoter(keyManager lh.KeyManager, blockHeight primitives.BlockHeight, view primitives.View, preparedMessages *lh.PreparedMessages) *VotesBuilder {
	voter := &Voter{keyManager, blockHeight, view, preparedMessages}
	builder.voters = append(builder.voters, voter)
	return builder
}

func NewVotesBuilder() *VotesBuilder {
	return &VotesBuilder{}
}

func ASimpleViewChangeVotes(membersKeyManagers []lh.KeyManager, blockHeight primitives.BlockHeight, view primitives.View) []*protocol.ViewChangeMessageContentBuilder {
	builder := NewVotesBuilder()
	for _, keyManager := range membersKeyManagers {
		builder.WithVoter(keyManager, blockHeight, view, nil)
	}
	return builder.Build()
}
