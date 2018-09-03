package electiontrigger

type ElectionTriggerMock struct {
	cb func()
}

func NewElectionTriggerMock() *ElectionTriggerMock {
	return &ElectionTriggerMock{}
}

func (tbet *ElectionTriggerMock) Start(cb func()) {
	tbet.cb = cb
}

func (tbet *ElectionTriggerMock) Stop() {
	tbet.cb = nil
}

func (tbet *ElectionTriggerMock) Trigger() {
	if tbet.cb != nil {
		tbet.cb()
	}
}
