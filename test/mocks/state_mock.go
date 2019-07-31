package mocks

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
)

type MockState struct {
	st state.State
}

func (s *MockState) SetHeightAndResetView(newHeight primitives.BlockHeight) *state.HeightView {
	return s.st.SetHeightAndResetView(newHeight)
}

func (s *MockState) SetView(newView primitives.View) *state.HeightView {
	return s.st.SetView(newView)
}

func (s *MockState) SetHeightView(newHeight primitives.BlockHeight, newView primitives.View) *state.HeightView {
	return s.st.SetHeightView(newHeight, newView)
}

func (s *MockState) Height() primitives.BlockHeight {
	return s.st.Height()
}

func (s *MockState) View() primitives.View {
	return s.st.View()
}

func (s *MockState) HeightView() *state.HeightView {
	return s.st.HeightView()
}

func NewMockState() *MockState {
	return &MockState{st: state.NewState()}
}

func (s *MockState) WithHeightView(h primitives.BlockHeight, v primitives.View) *MockState {
	s.st.SetHeightView(h, v)
	return s
}
