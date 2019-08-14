package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
)

type MockState struct {
	*state.State
}

func NewMockState() *MockState {
	return &MockState{State: state.NewState()}
}

func (s *MockState) WithHeightView(h primitives.BlockHeight, v primitives.View) *MockState {
	s.SetHeightView(context.Background(), h, v)
	return s
}
