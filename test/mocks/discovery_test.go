package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestDiscovery(t *testing.T) {
	genMemberId := func() primitives.MemberId {
		return primitives.MemberId(strconv.Itoa(rand.Int()))
	}

	t.Run("create a Discovery instance", func(t *testing.T) {
		instance := NewDiscovery()
		require.NotNil(t, instance, "Discovery instance created")
	})

	t.Run("get CommunicationMock instance by ID", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			id := genMemberId()
			gd := NewDiscovery()
			expectedCommunication := NewCommunication(gd)
			gd.RegisterCommunication(id, expectedCommunication)
			actualCommunication := gd.GetCommunicationById(id)
			require.Equal(t, expectedCommunication, actualCommunication, "received CommunicationMock instance by ID")
		})
	})

	t.Run("get all CommunicationMock IDs", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			id1 := genMemberId()
			id2 := genMemberId()
			id3 := genMemberId()
			gd := NewDiscovery()
			g1 := NewCommunication(gd)
			g2 := NewCommunication(gd)
			g3 := NewCommunication(gd)
			gd.RegisterCommunication(id1, g1)
			gd.RegisterCommunication(id2, g2)
			gd.RegisterCommunication(id3, g3)
			expectedMemberIdStrings := []string{id1.String(), id2.String(), id3.String()}
			actualMemberIds := gd.AllCommunicationsMemberIds()
			actualMemberIdStrings := make([]string, 0, len(actualMemberIds))
			for _, memberId := range actualMemberIds {
				actualMemberIdStrings = append(actualMemberIdStrings, memberId.String())
			}

			require.ElementsMatch(t, actualMemberIdStrings, expectedMemberIdStrings)
		})
	})

	t.Run("return communication=nil if given Id was not registered", func(t *testing.T) {
		id := genMemberId()
		gd := NewDiscovery()
		communication := gd.GetCommunicationById(id)

		require.Nil(t, communication, "GetCommunicationById() returns ok=false if ID not registered")
	})

	t.Run("return a list of all communications", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			gd := NewDiscovery()
			id1 := genMemberId()
			id2 := genMemberId()
			g1 := NewCommunication(gd)
			g2 := NewCommunication(gd)
			gd.RegisterCommunication(id1, g1)
			gd.RegisterCommunication(id2, g2)
			actual := gd.Communications(nil)
			expected := []*CommunicationMock{g1, g2}
			require.ElementsMatch(t, actual, expected, "list of all communications")
		})
	})

	t.Run("return a list of requested communications", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			gd := NewDiscovery()
			id1 := genMemberId()
			id2 := genMemberId()
			id3 := genMemberId()
			g1 := NewCommunication(gd)
			g2 := NewCommunication(gd)
			g3 := NewCommunication(gd)
			gd.RegisterCommunication(id1, g1)
			gd.RegisterCommunication(id2, g2)
			gd.RegisterCommunication(id3, g3)
			actual := gd.Communications([]primitives.MemberId{id1, id3})
			expected := []*CommunicationMock{g1, g3}
			require.ElementsMatch(t, actual, expected, "list of requested communications")
		})
	})

}
