package leaderelection

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

//func TestBlockchainGenerator(t *testing.T) {
//	t.Skip()
//
//	block1 := mocks.ABlock(interfaces.GenesisBlock)
//	block2 := mocks.ABlock(block1)
//	block3 := mocks.ABlock(block2)
//	block4 := mocks.ABlock(block3)
//
//	bc, err := GenerateBlockChainFor([]interfaces.Block{block1, block2, block3, block4})
//	if err != nil {
//		t.Fatalf("Error creating mock blockchain for tests: %s", err)
//		return
//	}
//
//	require.True(t, bc.LastBlock().Height() == primitives.BlockHeight(4))
//}

// TODO FLAKY
func TestGenerateProofsForTest(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	net := network.ABasicTestNetwork(ctx)

	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(block1)
	block3 := mocks.ABlock(block2)
	block4 := mocks.ABlock(block3)

	bc, err := GenerateProofsForTest([]interfaces.Block{block1, block2, block3, block4}, net.Nodes)

	if err != nil {
		t.Fatalf("Error creating mock blockchain for tests: %s", err)
		return
	}

	require.True(t, bc.LastBlock().Height() == primitives.BlockHeight(4))

}
