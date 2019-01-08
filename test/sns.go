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
