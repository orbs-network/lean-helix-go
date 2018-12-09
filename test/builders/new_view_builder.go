package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
)

type NewViewBuilder struct {
	leaderKeyManager   lh.KeyManager
	membersKeyManagers []lh.KeyManager
	block              lh.Block
	blockHeight        primitives.BlockHeight
	view               primitives.View
}

func (builder *NewViewBuilder) Build() *lh.NewViewMessage {
	ppmFactory := lh.NewMessageFactory(builder.leaderKeyManager)
	ppmCB := ppmFactory.CreatePreprepareMessageContentBuilder(builder.blockHeight, builder.view, builder.block, CalculateBlockHash(builder.block))

	var votes []*lh.ViewChangeMessageContentBuilder
	for _, keyManager := range builder.membersKeyManagers {
		messageFactory := lh.NewMessageFactory(keyManager)
		vcmCB := messageFactory.CreateViewChangeMessageContentBuilder(builder.blockHeight, builder.view, nil)
		votes = append(votes, vcmCB)
	}

	messageFactory := lh.NewMessageFactory(builder.leaderKeyManager)
	nvcb := messageFactory.CreateNewViewMessageContentBuilder(builder.blockHeight, builder.view, ppmCB, votes)
	return lh.NewNewViewMessage(nvcb.Build(), builder.block)
}

func (builder *NewViewBuilder) LeadBy(keyManager lh.KeyManager) *NewViewBuilder {
	builder.leaderKeyManager = keyManager
	return builder
}

func (builder *NewViewBuilder) WithMembers(membersKeyManagers []lh.KeyManager) *NewViewBuilder {
	builder.membersKeyManagers = membersKeyManagers
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
	return &NewViewBuilder{}
}
