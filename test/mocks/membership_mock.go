// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sort"
)

type FakeMembership struct {
	myMemberId             primitives.MemberId
	discovery              *Discovery
	orderCommitteeByHeight bool
}

func NewFakeMembership(myMemberId primitives.MemberId, discovery *Discovery, orderCommitteeByHeight bool) *FakeMembership {
	return &FakeMembership{
		myMemberId:             myMemberId,
		discovery:              discovery,
		orderCommitteeByHeight: orderCommitteeByHeight,
	}
}

func (m *FakeMembership) MyMemberId() primitives.MemberId {
	return m.myMemberId
}

func (m *FakeMembership) RequestOrderedCommittee(ctx context.Context, blockHeight primitives.BlockHeight, randomSeed uint64, prevBlockReferenceTime primitives.TimestampSeconds) ([]primitives.MemberId, error) {
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

func (m *FakeMembership) RequestCommitteeForBlockProof(ctx context.Context, prevBlockReferenceTime primitives.TimestampSeconds) ([]primitives.MemberId, error) {
	result := m.discovery.AllCommunicationsMemberIds()
	return result, nil
}
