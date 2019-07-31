package state

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sync"
)

// Mutable, goroutine-safe state object
type State interface {
	SetHeightAndResetView(newHeight primitives.BlockHeight) *HeightView
	SetView(newView primitives.View) *HeightView
	SetHeightView(newHeight primitives.BlockHeight, newView primitives.View) *HeightView
	Height() primitives.BlockHeight
	View() primitives.View
	HeightView() *HeightView
}

type state struct {
	sync.RWMutex
	height primitives.BlockHeight
	view   primitives.View
}

func (s *state) SetHeightAndResetView(newHeight primitives.BlockHeight) *HeightView {
	s.Lock()
	defer s.Unlock()

	s.height = newHeight
	s.view = 0
	return NewHeightView(s.height, s.view)
}

func (s *state) SetView(newView primitives.View) *HeightView {
	s.Lock()
	defer s.Unlock()

	s.view = newView
	return NewHeightView(s.height, s.view)
}

func (s *state) SetHeightView(newHeight primitives.BlockHeight, newView primitives.View) *HeightView {
	s.Lock()
	defer s.Unlock()

	s.height = newHeight
	s.view = newView
	return NewHeightView(s.height, s.view)
}

func (s *state) Height() primitives.BlockHeight {
	s.RLock()
	defer s.RUnlock()

	return s.height
}

func (s *state) View() primitives.View {
	s.RLock()
	defer s.RUnlock()

	return s.view
}

func (s *state) HeightView() *HeightView {
	s.RLock()
	defer s.RUnlock()

	return NewHeightView(s.height, s.view)
}

func NewState() State {
	return &state{
		height: 0,
		view:   0,
	}
}

// Immutable instance of height+view
type HeightView struct {
	height primitives.BlockHeight
	view   primitives.View
}

func (hv *HeightView) String() string {
	return fmt.Sprintf("H=%d,V=%d", hv.height, hv.view)
}

func NewHeightView(h primitives.BlockHeight, v primitives.View) *HeightView {
	return &HeightView{
		height: h,
		view:   v,
	}
}

func (hv *HeightView) Height() primitives.BlockHeight {
	return hv.height
}

func (hv *HeightView) View() primitives.View {
	return hv.view
}
