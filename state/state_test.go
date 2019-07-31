package state

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetHeightResetsView(t *testing.T) {
	state := NewState()
	state.SetHeightView(2, 3)
	require.Equal(t, primitives.BlockHeight(2), state.Height(), "returned incorrect height")
	require.Equal(t, primitives.View(3), state.View(), "returned incorrect view")
	state.SetHeightAndResetView(4)
	require.Equal(t, primitives.BlockHeight(4), state.Height(), "returned incorrect height")
	require.Equal(t, primitives.View(0), state.View(), "returned incorrect view")
}

func TestReturnCorrectHeightViewInstance(t *testing.T) {
	state := NewState()
	state.SetHeightView(2, 3)
	hv := state.HeightView()
	state.SetHeightView(4, 5)
	require.Equal(t, primitives.BlockHeight(2), hv.Height(), "returned incorrect height")
	require.Equal(t, primitives.View(3), hv.View(), "returned incorrect view")
}
