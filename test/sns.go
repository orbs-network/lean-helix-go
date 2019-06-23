// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import "context"

type Sns struct {
	signalChannel chan bool
	resumeChannel chan bool
}

func NewSignalAndStop() *Sns {
	return &Sns{
		signalChannel: make(chan bool),
		resumeChannel: make(chan bool),
	}
}

func (s *Sns) SignalAndStop(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case s.signalChannel <- true:
	}

	select {
	case <-ctx.Done():
		return
	case <-s.resumeChannel:
	}
}

func (s *Sns) WaitForSignal(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-s.signalChannel:
	}
}

func (s *Sns) Resume(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case s.resumeChannel <- true:
	}
}
