// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"bytes"
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRandomSeedToBytes(t *testing.T) {
	require.Equal(t, []byte{49}, randomseed.RandomSeedToBytes(1))
	require.Equal(t, []byte{49, 48, 48, 48}, randomseed.RandomSeedToBytes(1000))
	require.Equal(t, []byte{49, 50, 51, 52}, randomseed.RandomSeedToBytes(1234))
}

func TestCalculateNilRandomSeed(t *testing.T) {
	require.Equal(t, uint64(0x1b4ce424c81442e3), randomseed.CalculateRandomSeed(nil))
}

func TestCalculateRandomSeed(t *testing.T) {
	require.Equal(t, uint64(0xa228ab770a49c603), randomseed.CalculateRandomSeed([]byte{1, 2, 3}))
	require.Equal(t, uint64(0xbd0c0c86df41c870), randomseed.CalculateRandomSeed([]byte{0, 0, 0}))
	require.Equal(t, uint64(0xeae735db4cebd731), randomseed.CalculateRandomSeed([]byte{6, 6, 6, 6, 6, 6}))
}

func TestValidateRandomSeed(t *testing.T) {
	prevBlockProof := (&protocol.BlockProofBuilder{
		RandomSeedSignature: []byte{1, 2, 3},
	}).Build()

	blockProof := (&protocol.BlockProofBuilder{
		RandomSeedSignature: []byte{4, 5, 6},
	}).Build()

	memberId := primitives.MemberId("Dummy Member Id")
	keyManager := mocks.NewMockKeyManager(memberId)

	randomseed.ValidateRandomSeed(keyManager, 4, blockProof, prevBlockProof)

	randomSeedSignature := primitives.Signature(blockProof.RandomSeedSignature())
	prevRandomSeedSignature := prevBlockProof.RandomSeedSignature()

	randomSeed := randomseed.CalculateRandomSeed(prevRandomSeedSignature) // Calculate the random seed based on prev block proof

	lastCall := keyManager.VerifyRandomSeedHistory(0)
	require.Equal(t, primitives.BlockHeight(4), lastCall.BlockHeight)
	require.True(t, bytes.Equal(keyManager.VerifyRandomSeedHistory(0).Sender.Signature(), randomSeedSignature))
	require.True(t, bytes.Equal(keyManager.VerifyRandomSeedHistory(0).Content, randomseed.RandomSeedToBytes(randomSeed)))
}
