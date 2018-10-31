package test

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

// This file is based on PBFTTerm.spec.ts

const NODE_COUNT = 4

func TestViewIncrementedAfterElectionTrigger(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	net := builders.ATestNetwork(NODE_COUNT, nil)
	termConfig := lh.BuildTermConfig(net.Nodes[0].BuildConfig())
	term, err := lh.NewLeanHelixTerm(ctx, termConfig, 0, func(block lh.Block) {})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, View(0), term.GetView(), "Term should have view=0 on init")
	net.TriggerElection(ctx)
	require.Equal(t, View(1), term.GetView(), "Term should have view=1 after one election")
}

func TestRejectNewViewMessagesFromPast(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	height := BlockHeight(0)
	view := View(0)
	block := builders.CreateBlock(builders.GenesisBlock)
	net := builders.ATestNetwork(NODE_COUNT, nil)

	node := net.Nodes[0]
	messageFactory := lh.NewMessageFactory(node.KeyManager)
	ppmContentBuilder := messageFactory.CreatePreprepareMessageContentBuilder(height, view, block)
	nvm := messageFactory.CreateNewViewMessage(height, view, ppmContentBuilder, nil, block)
	termConfig := lh.BuildTermConfig(node.BuildConfig())
	term, err := lh.NewLeanHelixTerm(ctx, termConfig, height, func(block lh.Block) {})
	if err != nil {
		t.Fatal(err)
	}

	//require.Equal(t, View(0), term.GetView(), "Term should have view=0 on init")
	net.TriggerElection(ctx)
	//require.Equal(t, View(1), term.GetView(), "Term should have view=1 after one election")
	term.OnReceiveNewView(ctx, nvm)
	require.Equal(t, View(1), term.GetView(), "Term should have view=1")
}
