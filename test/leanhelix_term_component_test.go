package test

import (
	"context"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestViewIncrementedAfterElectionTrigger(t *testing.T) {
	net := builders.ABasicTestNetwork()
	termConfig := net.Nodes[0].BuildConfig()
	filter := lh.NewConsensusMessageFilter(termConfig.KeyManager.MyPublicKey())
	term := lh.NewLeanHelixTerm(termConfig, filter, 0)

	require.Equal(t, View(0), term.GetView(), "Term should have view=0 on init")
	net.TriggerElection()
	require.Equal(t, View(1), term.GetView(), "Term should have view=1 after one election")
}

func TestRejectNewViewMessagesFromPast(t *testing.T) {
	ctx := context.Background()

	height := BlockHeight(0)
	view := View(0)
	block := builders.CreateBlock(builders.GenesisBlock)
	net := builders.ABasicTestNetwork()

	node := net.Nodes[0]
	messageFactory := lh.NewMessageFactory(node.KeyManager)
	ppmContentBuilder := messageFactory.CreatePreprepareMessageContentBuilder(height, view, block)
	nvm := builders.ANewViewMessage(node.KeyManager, height, view, ppmContentBuilder, nil, block)
	termConfig := net.Nodes[0].BuildConfig()
	filter := lh.NewConsensusMessageFilter(termConfig.KeyManager.MyPublicKey())
	term := lh.NewLeanHelixTerm(termConfig, filter, 0)

	require.Equal(t, View(0), term.GetView(), "Term should have view=0 on init")
	net.TriggerElection()
	require.Equal(t, View(1), term.GetView(), "Term should have view=1 after one election")
	term.OnReceiveNewView(ctx, nvm)
	require.Equal(t, View(1), term.GetView(), "Term should have view=1")
}
