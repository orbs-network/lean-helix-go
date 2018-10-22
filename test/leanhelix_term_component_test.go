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

func triggerElection(testnet *builders.TestNetwork) {
	for _, node := range testnet.Nodes {
		node.TriggerElection()
	}
}

// Based on

// Based on "onReceivePrePrepare should accept views that match its current view"
func TestAcceptPreprepareWithCurrentView(t *testing.T) {

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
	triggerElection(net)
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
