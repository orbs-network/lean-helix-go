package mocks

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
)

type MockState struct {
	st state.State
}

func NewMockState() *MockState {
	return &MockState{st: state.NewState()}
}

func (s *MockState) WithHeightView(h primitives.BlockHeight, v primitives.View) *MockState {
	s.st.SetHeightView(h, v)
	return s
}
