// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

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
