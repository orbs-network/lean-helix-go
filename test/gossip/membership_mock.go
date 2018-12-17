package gossip

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sort"
)

type MockMembership struct {
	myMemberId primitives.MemberId
	discovery  *Discovery
}

func NewMockMembership(myMemberId primitives.MemberId, discovery *Discovery) *MockMembership {
	return &MockMembership{
		myMemberId: myMemberId,
		discovery:  discovery,
	}
}

func (m *MockMembership) MyMemberId() primitives.MemberId {
	return m.myMemberId
}

func (m *MockMembership) RequestOrderedCommittee(ctx context.Context, blockHeight primitives.BlockHeight, seed uint64, maxCommitteeSize uint32) []primitives.MemberId {
	result := m.discovery.AllGossipsMemberIds()
	sort.Slice(result, func(i, j int) bool {
		return result[i].KeyForMap() < result[j].KeyForMap()
	})
	return result
}
