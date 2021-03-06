package state

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetHeightResetsView(t *testing.T) {
	state := NewState()
	_, _ = state.SetHeightView(context.Background(), 2, 3)
	require.Equal(t, primitives.BlockHeight(2), state.Height(), "returned incorrect height")
	require.Equal(t, primitives.View(3), state.View(), "returned incorrect view")
	_, _ = state.SetHeightAndResetView(4)
	require.Equal(t, primitives.BlockHeight(4), state.Height(), "returned incorrect height")
	require.Equal(t, primitives.View(0), state.View(), "returned incorrect view")
}

func TestReturnCorrectHeightViewInstance(t *testing.T) {
	state := NewState()
	_, _ = state.SetHeightView(context.Background(), 2, 3)
	hv := state.HeightView()
	_, _ = state.SetHeightView(context.Background(), 4, 5)
	require.Equal(t, primitives.BlockHeight(2), hv.Height(), "returned incorrect height")
	require.Equal(t, primitives.View(3), hv.View(), "returned incorrect view")
}
