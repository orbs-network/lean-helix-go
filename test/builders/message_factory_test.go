package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	keyManager := NewMockKeyManager(lh.PublicKey("My PK"))
	term := lh.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := lh.View(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(GenesisBlock)
	//blockHash := CalculateBlockHash(MockBlock)
	fac := NewMockMessageFactory(CalculateBlockHash, keyManager)

	ppm := fac.CreatePreprepareMessage(term, view, block)

	require.Equal(t, term, ppm.SignedHeader().BlockHeight(), "expected height to be %s but got %s", term, ppm.SignedHeader().BlockHeight())

}
