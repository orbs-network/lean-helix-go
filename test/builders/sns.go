package builders

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

func (s *Sns) SignalAndStop() {
	s.signalChannel <- true
	<-s.resumeChannel
}

func (s *Sns) WaitForSignal() {
	<-s.signalChannel
}

func (s *Sns) Resume() {
	s.resumeChannel <- true
}
