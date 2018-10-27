package test

import (
	"context"
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

// This file is based on PBFTTerm.spec.ts

const NODE_COUNT = 4

// Based on

func TestViewIncrementedAfterElectionTrigger(t *testing.T) {

	ctx, ctxCancel := context.WithCancel(context.Background())

	net := builders.NewTestNetworkBuilder(NODE_COUNT).
		WithContext(ctx, ctxCancel).
		Build()

	termConfig := lh.BuildTermConfig(net.Nodes[0].Config)
	term, err := lh.NewLeanHelixTerm(ctx, termConfig, 0, func(block lh.Block) {})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, View(0), term.GetView(), "Term should have view=0 on init")
	net.TriggerElection(ctx)
	require.Equal(t, View(1), term.GetView(), "Term should have view=1 after one election")
}

func TestRejectNewViewMessagesFromPast(t *testing.T) {

	t.Skip() // this is stuck

	ctx, ctxCancel := context.WithCancel(context.Background())

	height := BlockHeight(0)
	view := View(0)
	block := builders.CreateBlock(builders.GenesisBlock)
	net := builders.NewTestNetworkBuilder(NODE_COUNT).
		WithContext(ctx, ctxCancel).
		Build()

	node := net.Nodes[0]
	node.Gossip.When("SendMessage", mock.Any, mock.Any, mock.Any).Return()
	messageFactory := lh.NewMessageFactory(node.KeyManager)
	ppmContentBuilder := messageFactory.CreatePreprepareMessageContentBuilder(height, view, block)
	nvm := messageFactory.CreateNewViewMessage(height, view, ppmContentBuilder, nil, block)
	termConfig := lh.BuildTermConfig(node.Config)
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

// Based on "onReceivePrePrepare should accept views that match its current view"
func TestAcceptPreprepareWithCurrentView(t *testing.T) {

	t.Skip()

	ctx, ctxCancel := context.WithCancel(context.Background())

	net := builders.NewTestNetworkBuilder(NODE_COUNT).
		WithContext(ctx, ctxCancel).
		Build()
	node1 := net.Nodes[1]
	termConfig1 := lh.BuildTermConfig(node1.Config)
	mockStorage1 := builders.NewMockStorage()

	// TODO This is smelly - maybe wait till correct config architecture
	// emerges from future tests
	termConfig1.Storage = mockStorage1
	node1LeanHelixTerm, err := lh.NewLeanHelixTerm(ctx, termConfig1, 0, nil)
	if err != nil {
		t.Error(err)
	}
	require.Equal(t, node1LeanHelixTerm.GetView(), View(0), "Node 1 view should be 0")
	net.TriggerElection(ctx)
	require.Equal(t, node1LeanHelixTerm.GetView(), View(1), "Node 1 view should be 1")

	block := builders.CreateBlock(builders.GenesisBlock)
	// spy on storePrepare

	keyManager := node1.Config.KeyManager
	utils := node1.Config.BlockUtils.(*builders.MockBlockUtils)
	mf1 := lh.NewMessageFactory(keyManager)

	ppmFromCurrentView := mf1.CreatePreprepareMessage(1, 1, block)
	node1LeanHelixTerm.OnReceivePreprepare(ctx, ppmFromCurrentView)
	utils.ResolveAllValidations(true)
	mockStorage1.When("StorePrepare").Times(1)
	mockStorage1.Verify()
	mockStorage1.Reset()

	ppmFromFutureView := mf1.CreatePreprepareMessage(1, 2, block)
	node1LeanHelixTerm.OnReceivePreprepare(ctx, ppmFromFutureView)
	utils.ResolveAllValidations(true)
	mockStorage1.Never("StorePrepare")
	mockStorage1.Verify()
	mockStorage1.Reset()

	ppmFromPastView := mf1.CreatePreprepareMessage(1, 0, block)
	node1LeanHelixTerm.OnReceivePreprepare(ctx, ppmFromPastView)
	utils.ResolveAllValidations(true)
	mockStorage1.Never("StorePrepare")
	mockStorage1.Verify()
	mockStorage1.Reset()
}
