package test

import (
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	blockChain := builders.NewInMemoryBlockChain()
	actual := blockChain.GetLastBlock()
	expected := builders.GenesisBlock
	require.Equal(t, expected, actual, "Did not return the genesis block as the first block")
}

func TestAppendingToBlockChain(t *testing.T) {
	blockChain := builders.NewInMemoryBlockChain()
	block := builders.CreateBlock(builders.GenesisBlock)
	blockChain.AppendBlockToChain(block)

	actual := blockChain.GetLastBlock()
	require.Equal(t, block, actual, "Did not return the genesis block as the first block")
}

func TestGettingTheLatestBlock(t *testing.T) {
	blockChain := builders.NewInMemoryBlockChain()
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(block1)
	block3 := builders.CreateBlock(block2)
	blockChain.AppendBlockToChain(block1)
	blockChain.AppendBlockToChain(block2)
	blockChain.AppendBlockToChain(block3)

	actual := blockChain.GetLastBlock()
	require.Equal(t, block3, actual, "Did not return the latest block")
}
