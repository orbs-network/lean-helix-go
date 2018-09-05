package gossip

import (
	"github.com/orbs-network/lean-helix-go/types"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestGossipDiscovery(t *testing.T) {
	genId := func() types.PublicKey {
		return types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	}

	t.Run("create a GossipDiscovery instance", func(t *testing.T) {
		instance := NewGossipDiscovery()
		require.NotNil(t, instance, "GossipDiscovery instance created")
	})

	t.Run("get Gossip instance by ID", func(t *testing.T) {
		id := genId()
		gd := NewGossipDiscovery()
		expectedGossip := NewGossip(gd)
		gd.RegisterGossip(id, expectedGossip)
		actualGossip, _ := gd.GetGossipByPK(id)
		require.Equal(t, expectedGossip, actualGossip, "received Gossip instance by ID")
	})

	t.Run("get all Gossip IDs", func(t *testing.T) {
		id1 := genId()
		id2 := genId()
		id3 := genId()
		gd := NewGossipDiscovery()
		g1 := NewGossip(gd)
		g2 := NewGossip(gd)
		g3 := NewGossip(gd)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		gd.RegisterGossip(id3, g3)
		expected := []types.PublicKey{id1, id2, id3}
		actual := gd.getAllGossipsPKs()

		require.ElementsMatch(t, actual, expected)
	})

	t.Run("return ok=false if given Id was not registered", func(t *testing.T) {
		id := genId()
		gd := NewGossipDiscovery()
		_, ok := gd.GetGossipByPK(id)

		require.False(t, ok, "GetGossipByPK() returns ok=false if ID not registered")
	})

	t.Run("return a list of all gossips", func(t *testing.T) {
		gd := NewGossipDiscovery()
		id1 := genId()
		id2 := genId()
		g1 := NewGossip(gd)
		g2 := NewGossip(gd)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		actual := gd.GetGossips(nil)
		expected := []*Gossip{g1, g2}
		require.ElementsMatch(t, actual, expected, "list of all gossips")
	})

	t.Run("return a list of requested gossips", func(t *testing.T) {
		gd := NewGossipDiscovery()
		id1 := genId()
		id2 := genId()
		id3 := genId()
		g1 := NewGossip(gd)
		g2 := NewGossip(gd)
		g3 := NewGossip(gd)
		gd.RegisterGossip(id1, g1)
		gd.RegisterGossip(id2, g2)
		gd.RegisterGossip(id3, g3)
		actual := gd.GetGossips([]types.PublicKey{id1, id3})
		expected := []*Gossip{g1, g3}
		require.ElementsMatch(t, actual, expected, "list of requested gossips")
	})

}
