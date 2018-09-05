package inmemoryblockchain_test

import (
	"github.com/orbs-network/lean-helix-go/go/test/builders"
	"github.com/orbs-network/lean-helix-go/go/test/inmemoryblockchain"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	blockChain := inmemoryblockchain.NewInMemoryBlockChain()
	actual := blockChain.GetLastBlock()
	expected := inmemoryblockchain.GenesisBlock
	require.Equal(t, expected, actual, "Did not return the genesis block as the first block")
}

func TestAppendingToBlockChain(t *testing.T) {
	blockChain := inmemoryblockchain.NewInMemoryBlockChain()
	block := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	blockChain.AppendBlockToChain(block)

	actual := blockChain.GetLastBlock()
	require.Equal(t, block, actual, "Did not return the genesis block as the first block")
}

func TestGettingTheLatestBlock(t *testing.T) {
	blockChain := inmemoryblockchain.NewInMemoryBlockChain()
	block1 := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	block2 := builders.CreateBlock(block1)
	block3 := builders.CreateBlock(block2)
	blockChain.AppendBlockToChain(block1)
	blockChain.AppendBlockToChain(block2)
	blockChain.AppendBlockToChain(block3)

	actual := blockChain.GetLastBlock()
	require.Equal(t, block3, actual, "Did not return the latest block")
}
