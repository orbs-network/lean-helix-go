package gossip

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestDiscovery(t *testing.T) {
	genMemberId := func() primitives.MemberId {
		return primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000000))))
	}

	t.Run("create a Discovery instance", func(t *testing.T) {
		instance := NewDiscovery()
		require.NotNil(t, instance, "Discovery instance created")
	})

	t.Run("get Gossip instance by ID", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			id := genMemberId()
			gd := NewDiscovery()
			expectedGossip := NewGossip(gd)
			gd.RegisterGossip(id, expectedGossip)
			actualGossip := gd.GetGossipById(id)
			require.Equal(t, expectedGossip, actualGossip, "received Gossip instance by ID")
		})
	})

	t.Run("get all Gossip IDs", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			id1 := genMemberId()
			id2 := genMemberId()
			id3 := genMemberId()
			gd := NewDiscovery()
			g1 := NewGossip(gd)
			g2 := NewGossip(gd)
			g3 := NewGossip(gd)
			gd.RegisterGossip(id1, g1)
			gd.RegisterGossip(id2, g2)
			gd.RegisterGossip(id3, g3)
			expectedMemberIdStrings := []string{id1.String(), id2.String(), id3.String()}
			actualMemberIds := gd.AllGossipsMemberIds()
			actualMemberIdStrings := make([]string, 0, len(actualMemberIds))
			for _, memberId := range actualMemberIds {
				actualMemberIdStrings = append(actualMemberIdStrings, memberId.String())
			}

			require.ElementsMatch(t, actualMemberIdStrings, expectedMemberIdStrings)
		})
	})

	t.Run("return gossip=nil if given Id was not registered", func(t *testing.T) {
		id := genMemberId()
		gd := NewDiscovery()
		gossip := gd.GetGossipById(id)

		require.Nil(t, gossip, "GetGossipById() returns ok=false if ID not registered")
	})

	t.Run("return a list of all gossips", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			gd := NewDiscovery()
			id1 := genMemberId()
			id2 := genMemberId()
			g1 := NewGossip(gd)
			g2 := NewGossip(gd)
			gd.RegisterGossip(id1, g1)
			gd.RegisterGossip(id2, g2)
			actual := gd.Gossips(nil)
			expected := []*Gossip{g1, g2}
			require.ElementsMatch(t, actual, expected, "list of all gossips")
		})
	})

	t.Run("return a list of requested gossips", func(t *testing.T) {
		test.WithContext(func(ctx context.Context) {
			gd := NewDiscovery()
			id1 := genMemberId()
			id2 := genMemberId()
			id3 := genMemberId()
			g1 := NewGossip(gd)
			g2 := NewGossip(gd)
			g3 := NewGossip(gd)
			gd.RegisterGossip(id1, g1)
			gd.RegisterGossip(id2, g2)
			gd.RegisterGossip(id3, g3)
			actual := gd.Gossips([]primitives.MemberId{id1, id3})
			expected := []*Gossip{g1, g3}
			require.ElementsMatch(t, actual, expected, "list of requested gossips")
		})
	})

}
