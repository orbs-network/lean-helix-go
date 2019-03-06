package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sort"
)

type MockMembership struct {
	myMemberId             primitives.MemberId
	discovery              *Discovery
	orderCommitteeByHeight bool
}

func NewMockMembership(myMemberId primitives.MemberId, discovery *Discovery, orderCommitteeByHeight bool) *MockMembership {
	return &MockMembership{
		myMemberId:             myMemberId,
		discovery:              discovery,
		orderCommitteeByHeight: orderCommitteeByHeight,
	}
}

func (m *MockMembership) MyMemberId() primitives.MemberId {
	return m.myMemberId
}

func (m *MockMembership) RequestOrderedCommittee(ctx context.Context, blockHeight primitives.BlockHeight, randomSeed uint64) ([]primitives.MemberId, error) {
	result := m.discovery.AllCommunicationsMemberIds()
	sort.Slice(result, func(i, j int) bool {
		return result[i].KeyForMap() < result[j].KeyForMap()
	})

	// we want to replace the leader every height,
	// we just shift all the ordered nodes according to the given height
	if m.orderCommitteeByHeight {
		for i := 0; i < int(blockHeight); i++ {
			result = append(result[1:], result[0]) // shift left (circular)
		}
	}

	return result, nil
}
