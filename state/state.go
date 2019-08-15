package state

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/pkg/errors"
	"sync"
)

// Mutable, goroutine-safe State object
type State struct {
	sync.RWMutex
	height primitives.BlockHeight
	view   primitives.View
}

func (s *State) CompareWithEffectiveHeightAndCancel(cancel context.CancelFunc, lastUpdated primitives.BlockHeight, height primitives.BlockHeight) (*HeightView, bool) {
	s.RLock()
	defer s.RUnlock()

	hv := NewHeightView(s.height, s.view)

	effectiveHeight := s.height
	if effectiveHeight < lastUpdated {
		effectiveHeight = lastUpdated
	}

	if height < effectiveHeight {
		return hv, false
	}

	cancel()
	return hv, true
}

func (s *State) CancelContextIfHeightViewUnchanged(cancel context.CancelFunc, height primitives.BlockHeight, view primitives.View) (*HeightView, bool) {
	s.RLock()
	defer s.RUnlock()

	hv := NewHeightView(s.height, s.view)

	if s.height != height || s.view != view {
		return hv, false
	}

	cancel()
	return hv, true
}

func (s *State) SetHeightAndResetView(ctx context.Context, newHeight primitives.BlockHeight) (*HeightView, error) {
	s.Lock()
	defer s.Unlock()

	if ctx.Err() == context.Canceled {
		return NewHeightView(s.height, s.view), ctx.Err()
	}

	if s.height >= newHeight {
		return NewHeightView(s.height, s.view), errors.New("SetHeightAndResetView() failed because newHeight is not newer than current height")
	}

	s.height = newHeight
	s.view = 0
	return NewHeightView(s.height, s.view), nil
}

func (s *State) SetView(ctx context.Context, newView primitives.View) (*HeightView, error) {
	s.Lock()
	defer s.Unlock()

	if ctx.Err() == context.Canceled {
		return NewHeightView(s.height, s.view), ctx.Err()
	}

	if s.view > newView && newView != 0 {
		return NewHeightView(s.height, s.view), errors.New("SetView() failed because newView is not newer than current view, and it's not a new term")
	}

	s.view = newView
	return NewHeightView(s.height, s.view), nil
}

// TODO For testing only, so perhaps move it away
func (s *State) SetHeightView(ctx context.Context, newHeight primitives.BlockHeight, newView primitives.View) (*HeightView, error) {
	s.Lock()
	defer s.Unlock()

	if ctx.Err() == context.Canceled {
		return NewHeightView(s.height, s.view), ctx.Err()
	}

	s.height = newHeight
	s.view = newView
	return NewHeightView(s.height, s.view), nil
}

func (s *State) Height() primitives.BlockHeight {
	s.RLock()
	defer s.RUnlock()

	return s.height
}

func (s *State) View() primitives.View {
	s.RLock()
	defer s.RUnlock()

	return s.view
}

func (s *State) HeightView() *HeightView {
	s.RLock()
	defer s.RUnlock()

	return NewHeightView(s.height, s.view)
}

func NewState() *State {
	return &State{
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
