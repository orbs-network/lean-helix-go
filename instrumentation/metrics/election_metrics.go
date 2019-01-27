package metrics

import "github.com/orbs-network/lean-helix-go/spec/types/go/primitives"

type ElectionMetrics interface {
	CurrentLeaderMemberId() primitives.MemberId
	CurrentView() primitives.View
}

type electionMetrics struct {
	currentLeaderMemberId primitives.MemberId
	currentView           primitives.View
}

func (m *electionMetrics) CurrentLeaderMemberId() primitives.MemberId {
	return m.currentLeaderMemberId
}

func (m *electionMetrics) CurrentView() primitives.View {
	return m.currentView
}

func NewElectionMetrics(currentLeaderMemberId primitives.MemberId, currentView primitives.View) ElectionMetrics {
	return &electionMetrics{
		currentLeaderMemberId: currentLeaderMemberId,
		currentView:           currentView,
	}
}
