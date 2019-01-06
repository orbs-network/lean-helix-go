package test

import (
	"github.com/orbs-network/lean-helix-go/services/randomseed"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRandomSeedToBytes(t *testing.T) {
	require.Equal(t, []byte{49}, randomseed.RandomSeedToBytes(1))
	require.Equal(t, []byte{49, 48, 48, 48}, randomseed.RandomSeedToBytes(1000))
	require.Equal(t, []byte{49, 50, 51, 52}, randomseed.RandomSeedToBytes(1234))
}
