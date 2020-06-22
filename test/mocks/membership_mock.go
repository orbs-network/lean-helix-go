// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
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

func (m *FakeMembership) RequestOrderedCommittee(ctx context.Context, blockHeight primitives.BlockHeight, randomSeed uint64, prevBlockReferenceTime primitives.TimestampSeconds) ([]interfaces.CommitteeMember, error) {
	memberIds := m.discovery.AllCommunicationsMemberIds()
	sort.Slice(memberIds, func(i, j int) bool {
		return memberIds[i].KeyForMap() < memberIds[j].KeyForMap()
	})

	committeeMembers := make([]interfaces.CommitteeMember, len(memberIds))
	for i := 0; i < len(committeeMembers); i++ {
		committeeMembers[i].Weight = 1 // todo configurable weight
		committeeMembers[i].Id = memberIds[i]
	}

	// we want to replace the leader every height,
	// we just shift all the ordered nodes according to the given height
	if m.orderCommitteeByHeight {
		for i := 0; i < int(blockHeight); i++ {
			committeeMembers = append(committeeMembers[1:], committeeMembers[0]) // shift left (circular)
		}
	}

	return committeeMembers, nil
}

func (m *FakeMembership) RequestCommitteeForBlockProof(ctx context.Context, prevBlockReferenceTime primitives.TimestampSeconds) ([]interfaces.CommitteeMember, error) {
	memberIds := m.discovery.AllCommunicationsMemberIds()

	committeeMembers := make([]interfaces.CommitteeMember, len(memberIds))
	for i := 0; i < len(committeeMembers); i++ {
		committeeMembers[i].Weight = 1 // todo configurable weight
		committeeMembers[i].Id = memberIds[i]
	}

	return committeeMembers, nil
}
