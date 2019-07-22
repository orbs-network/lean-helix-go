package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sync"
)

// Mutable, goroutine-safe state object
type State interface {
	SetHeight(newHeight primitives.BlockHeight) *HeightView
	SetView(newView primitives.View) *HeightView
	SetHeightView(newHeight primitives.BlockHeight, newView primitives.View) *HeightView
	Height() primitives.BlockHeight
	View() primitives.View
	HeightView() *HeightView
}

type state struct {
	lock   sync.RWMutex
	height primitives.BlockHeight
	view   primitives.View
}

func (s *state) SetHeight(newHeight primitives.BlockHeight) *HeightView {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.height = newHeight
	s.view = 0
	return &HeightView{
		height: s.height,
		view:   s.view,
	}
}

func (s *state) SetView(newView primitives.View) *HeightView {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.view = newView
	return &HeightView{
		height: s.height,
		view:   s.view,
	}
}

func (s *state) SetHeightView(newHeight primitives.BlockHeight, newView primitives.View) *HeightView {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.height = newHeight
	s.view = newView
	return &HeightView{
		height: s.height,
		view:   s.view,
	}
}

func (s *state) Height() primitives.BlockHeight {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.height
}

func (s *state) View() primitives.View {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.view
}

func (s *state) HeightView() *HeightView {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return &HeightView{
		height: s.height,
		view:   s.view,
	}
}

func NewState() *state {
	return &state{
		height: 0,
		view:   0,
	}
}

type HeightView struct {
	height primitives.BlockHeight
	view   primitives.View
}

func (hv *HeightView) Height() primitives.BlockHeight {
	return hv.height
}

func (hv *HeightView) View() primitives.View {
	return hv.view
}
