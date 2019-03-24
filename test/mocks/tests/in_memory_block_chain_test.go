// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package tests

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	blockChain := mocks.NewInMemoryBlockChain()
	actual := blockChain.GetLastBlock()
	expected := interfaces.GenesisBlock
	require.Equal(t, expected, actual, "Did not return the genesis block as the first block")
}

func TestAppendingToBlockChain(t *testing.T) {
	blockChain := mocks.NewInMemoryBlockChain()
	block := mocks.ABlock(interfaces.GenesisBlock)
	blockChain.AppendBlockToChain(block, nil)

	actual := blockChain.GetLastBlock()
	require.Equal(t, block, actual, "Did not return the genesis block as the first block")
}

func TestGettingTheLatestBlock(t *testing.T) {
	blockChain := mocks.NewInMemoryBlockChain()
	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(block1)
	block3 := mocks.ABlock(block2)
	blockChain.AppendBlockToChain(block1, nil)
	blockChain.AppendBlockToChain(block2, nil)
	blockChain.AppendBlockToChain(block3, nil)

	actual := blockChain.GetLastBlock()
	require.Equal(t, block3, actual, "Did not return the latest block")
}

func TestGettingTheLatestBlockProof(t *testing.T) {
	blockChain := mocks.NewInMemoryBlockChain()
	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(block1)
	block3 := mocks.ABlock(block2)
	blockChain.AppendBlockToChain(block1, []byte{1, 2, 3})
	blockChain.AppendBlockToChain(block2, []byte{4, 5, 6})
	blockChain.AppendBlockToChain(block3, []byte{7, 8, 9})

	actual := blockChain.GetLastBlockProof()
	require.Equal(t, []byte{7, 8, 9}, actual, "Did not return the latest block proof")
}
